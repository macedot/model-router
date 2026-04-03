package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"model-router/models"
	"model-router/services"
)

// NewAnthropicHandler returns an http.HandlerFunc for Anthropic messages.
func NewAnthropicHandler(registry services.RegistryReader, forwarder *services.Forwarder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.AnthropicRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":{"type":"invalid_request_error","message":"Invalid request body: `+err.Error()+`"}}`, http.StatusBadRequest)
			return
		}

		if req.Model == "" {
			http.Error(w, `{"error":{"type":"invalid_request_error","message":"model is required"}}`, http.StatusBadRequest)
			return
		}

		if len(req.Messages) == 0 {
			http.Error(w, `{"error":{"type":"invalid_request_error","message":"messages is required and cannot be empty"}}`, http.StatusBadRequest)
			return
		}

		internalModel := registry.Get(req.Model)
		if internalModel == nil {
			http.Error(w, `{"error":{"type":"invalid_request_error","message":"Model not found: `+req.Model+`"}}`, http.StatusNotFound)
			return
		}

		var lastErr error
		for i, external := range internalModel.Externals {
			respBody, err := forwarder.ForwardAnthropic(r.Context(), &req, external)
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
				time.Sleep(time.Duration(internalModel.RetryDelaySecs) * time.Second)
			}
		}

		http.Error(w, `{"error":{"type":"api_error","message":"All providers failed: `+lastErr.Error()+`"}}`, http.StatusBadGateway)
	}
}
