package handlers

import (
	"encoding/json"
	"net/http"

	"model-router/models"
	"model-router/services"
)

type modelResponse struct {
	Name           string               `json:"name"`
	RequestFormat  models.RequestFormat  `json:"request_format"`
	Strategy       models.Strategy       `json:"strategy"`
	RetryDelaySecs uint32               `json:"retry_delay_secs"`
	Externals      []externalResponse   `json:"externals"`
}

type externalResponse struct {
	Name   string              `json:"name"`
	Format models.RequestFormat `json:"format"`
}

type providerResponse struct {
	ID   string              `json:"id"`
	URL  string              `json:"url"`
	Type models.RequestFormat `json:"type"`
}

// NewModelsHandler returns an http.HandlerFunc for listing models.
func NewModelsHandler(registry services.RegistryReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		modelsList := registry.List()

		resp := make([]modelResponse, len(modelsList))
		for i, m := range modelsList {
			externals := make([]externalResponse, len(m.Externals))
			for j, e := range m.Externals {
				externals[j] = externalResponse{
					Name:   e.Name,
					Format: e.Format,
				}
			}
			resp[i] = modelResponse{
				Name:           m.Name,
				RequestFormat:  m.RequestFormat,
				Strategy:       m.Strategy,
				RetryDelaySecs: m.RetryDelaySecs,
				Externals:      externals,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		providers := registry.ListProviders()
		provResp := make([]providerResponse, len(providers))
		for i, p := range providers {
			provResp[i] = providerResponse{
				ID:   p.ID,
				URL:  p.URL,
				Type: p.Format,
			}
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"models": resp, "providers": provResp})
	}
}
