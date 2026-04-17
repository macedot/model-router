package handlers

import (
	"log"
	"net/http"
	"time"

	"model-router/models"
	"model-router/services"
)

// NewOpenAIHandler returns an http.HandlerFunc for OpenAI chat completions.
func NewOpenAIHandler(registry services.RegistryReader, forwarder *services.Forwarder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, envelope, err := readBody(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request body", "invalid_request_error", err)
			return
		}

		if envelope.Model == "" {
			http.Error(w, `{"error":{"message":"model is required","type":"invalid_request_error"}}`, http.StatusBadRequest)
			return
		}

		if !envelope.HasMessages() {
			http.Error(w, `{"error":{"message":"messages is required and cannot be empty","type":"invalid_request_error"}}`, http.StatusBadRequest)
			return
		}

		internalModel := registry.Get(envelope.Model)
		if internalModel == nil {
			provider := registry.GetProvider(envelope.Model)
			if provider != nil {
				ext := provider.ToExternal()
				internalModel = &models.InternalModel{
					Name:          ext.Name,
					RequestFormat: models.FormatOpenAI,
					Strategy:      models.StrategyFallback,
					Externals:     []models.ExternalModel{ext},
				}
			}
		}
		if internalModel == nil {
			http.Error(w, `{"error":{"message":"Model not found: `+envelope.Model+`","type":"invalid_request_error"}}`, http.StatusNotFound)
			return
		}

		// Check if streaming
		streaming := envelope.Stream != nil && *envelope.Stream

		if streaming {
			if len(internalModel.Externals) == 0 {
				http.Error(w, `{"error":{"message":"No external providers configured","type":"invalid_request_error"}}`, http.StatusInternalServerError)
				return
			}
			external := internalModel.Externals[0]
			reqBody, err := services.PrepareRequest(body, external.Name, models.FormatOpenAI, external.Format)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "Failed to prepare request", "api_error", err)
				return
			}
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Transfer-Encoding", "chunked")
			if err := forwarder.ForwardStream(r.Context(), reqBody, external, w); err != nil {
				log.Printf("[handlers] stream forward error: %v", err)
			}
			return
		}

		// Non-streaming path supports fallback retry
		var lastErr error
		for i, external := range internalModel.Externals {
			reqBody, err := services.PrepareRequest(body, external.Name, models.FormatOpenAI, external.Format)
			if err != nil {
				lastErr = err
				continue
			}
			respBody, err := forwarder.Forward(r.Context(), reqBody, external)
			if err == nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write(respBody)
				return
			}
			lastErr = err

			// Retry delay between externals (skip delay after last one)
			if internalModel.Strategy == models.StrategyFallback &&
				i < len(internalModel.Externals)-1 &&
				internalModel.RetryDelaySecs > 0 {
				select {
				case <-r.Context().Done():
					return
				case <-time.After(time.Duration(internalModel.RetryDelaySecs) * time.Second):
				}
			}
		}

		writeError(w, http.StatusBadGateway, "All providers failed", "api_error", lastErr)
	}
}
