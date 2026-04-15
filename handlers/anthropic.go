package handlers

import (
	"net/http"
	"time"

	"model-router/models"
	"model-router/services"
)

// NewAnthropicHandler returns an http.HandlerFunc for Anthropic messages.
func NewAnthropicHandler(registry services.RegistryReader, forwarder *services.Forwarder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, envelope, err := readBody(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request body", "invalid_request_error", err)
			return
		}

		if envelope.Model == "" {
			http.Error(w, `{"error":{"type":"invalid_request_error","message":"model is required"}}`, http.StatusBadRequest)
			return
		}

		if !envelope.HasMessages() {
			http.Error(w, `{"error":{"type":"invalid_request_error","message":"messages is required and cannot be empty"}}`, http.StatusBadRequest)
			return
		}

		internalModel := registry.Get(envelope.Model)
		if internalModel == nil {
			http.Error(w, `{"error":{"type":"invalid_request_error","message":"Model not found: `+envelope.Model+`"}}`, http.StatusNotFound)
			return
		}

		var lastErr error
		for i, external := range internalModel.Externals {
			reqBody, err := services.PrepareRequest(body, external.Name, models.FormatAnthropic, external.Format)
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
