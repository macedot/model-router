package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"model-router/models"
	"model-router/services"
)

func TestModelsHandler_List(t *testing.T) {
	registry := services.NewRegistry([]models.InternalModel{
		{
			Name:          "model-a",
			RequestFormat: models.FormatOpenAI,
			Strategy:      models.StrategyFallback,
			Externals:     []models.ExternalModel{{Name: "ext-a"}},
		},
		{
			Name:          "model-b",
			RequestFormat: models.FormatAnthropic,
			Strategy:      models.StrategyFallback,
			Externals:     []models.ExternalModel{{Name: "ext-b"}},
		},
	})
	handler := NewModelsHandler(registry)

	mux := http.NewServeMux()
	mux.HandleFunc("/models", handler)

	req := httptest.NewRequest(http.MethodGet, "/models", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusOK)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	modelsList, ok := result["models"].([]interface{})
	if !ok {
		t.Fatal("expected models field to be an array")
	}
	if len(modelsList) != 2 {
		t.Errorf("len(models) = %d; want 2", len(modelsList))
	}
}

func TestModelsHandler_List_Empty(t *testing.T) {
	registry := services.NewRegistry([]models.InternalModel{})
	handler := NewModelsHandler(registry)

	mux := http.NewServeMux()
	mux.HandleFunc("/models", handler)

	req := httptest.NewRequest(http.MethodGet, "/models", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusOK)
	}

	var result map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&result)

	modelsList, ok := result["models"].([]interface{})
	if !ok {
		t.Fatal("expected models field to be an array")
	}
	if len(modelsList) != 0 {
		t.Errorf("len(models) = %d; want 0", len(modelsList))
	}
}
