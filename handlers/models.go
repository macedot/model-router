package handlers

import (
	"encoding/json"
	"net/http"

	"model-router/services"
)

// NewModelsHandler returns an http.HandlerFunc for listing models.
func NewModelsHandler(registry services.RegistryReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		modelsList := registry.List()

		names := make([]string, len(modelsList))
		for i, m := range modelsList {
			names[i] = m.Name
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"models": names})
	}
}
