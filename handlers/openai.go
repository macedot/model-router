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
		var req models.OpenAIRequest
		if err := decodeWithLimit(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request body", "invalid_request_error", err)
			return
		}

		if req.Model == "" {
			http.Error(w, `{"error":{"message":"model is required","type":"invalid_request_error"}}`, http.StatusBadRequest)
			return
		}

		if len(req.Messages) == 0 {
			http.Error(w, `{"error":{"message":"messages is required and cannot be empty","type":"invalid_request_error"}}`, http.StatusBadRequest)
			return
		}

		internalModel := registry.Get(req.Model)
		if internalModel == nil {
			http.Error(w, `{"error":{"message":"Model not found: `+req.Model+`","type":"invalid_request_error"}}`, http.StatusNotFound)
			return
		}

		// Check if streaming
		streaming := req.Stream != nil && *req.Stream

		if streaming {
			if len(internalModel.Externals) == 0 {
				http.Error(w, `{"error":{"message":"No external providers configured","type":"invalid_request_error"}}`, http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Transfer-Encoding", "chunked")
			// ForwardOpenAIStream writes directly to w with flushWriter (goproxy pattern)
			if err := forwarder.ForwardOpenAIStream(r.Context(), &req, internalModel.Externals[0], w); err != nil {
				// Headers already sent; log but can't write error to client mid-stream
				log.Printf("[handlers] stream forward error: %v", err)
			}
			return
		}

		// Non-streaming path supports fallback retry
		var lastErr error
		for i, external := range internalModel.Externals {
			respBody, err := forwarder.ForwardOpenAI(r.Context(), &req, external)
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
