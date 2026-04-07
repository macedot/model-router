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
			Name:           "model-a",
			RequestFormat:  models.FormatOpenAI,
			Strategy:       models.StrategyFallback,
			RetryDelaySecs: 2,
			Externals: []models.ExternalModel{
				{Name: "ext-a", URL: "https://secret.example.com", APIKey: "secret-key", Format: models.FormatOpenAI},
			},
		},
		{
			Name:          "model-b",
			RequestFormat: models.FormatAnthropic,
			Strategy:      models.StrategyFallback,
			Externals: []models.ExternalModel{
				{Name: "ext-b", Format: models.FormatAnthropic},
			},
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

	var result struct {
		Models []modelResponse `json:"models"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(result.Models) != 2 {
		t.Fatalf("len(models) = %d; want 2", len(result.Models))
	}

	// Verify first model details
	m := result.Models[0]
	if m.Name != "model-a" {
		t.Errorf("name = %q; want %q", m.Name, "model-a")
	}
	if m.RequestFormat != models.FormatOpenAI {
		t.Errorf("request_format = %q; want %q", m.RequestFormat, models.FormatOpenAI)
	}
	if m.Strategy != models.StrategyFallback {
		t.Errorf("strategy = %q; want %q", m.Strategy, models.StrategyFallback)
	}
	if m.RetryDelaySecs != 2 {
		t.Errorf("retry_delay_secs = %d; want 2", m.RetryDelaySecs)
	}
	if len(m.Externals) != 1 {
		t.Fatalf("len(externals) = %d; want 1", len(m.Externals))
	}
	if m.Externals[0].Name != "ext-a" {
		t.Errorf("external name = %q; want %q", m.Externals[0].Name, "ext-a")
	}
	if m.Externals[0].Format != models.FormatOpenAI {
		t.Errorf("external format = %q; want %q", m.Externals[0].Format, models.FormatOpenAI)
	}

	// Verify sensitive fields are not exposed in raw JSON
	body := rec.Body.String()
	if containsSensitive(body) {
		t.Error("response should not contain sensitive fields (url, api_key)")
	}
}

// containsSensitive checks that no external URL or API key leaked into the response.
func containsSensitive(body string) bool {
	return json.Unmarshal([]byte(body), &struct{}{}) == nil &&
		(containsString(body, "secret.example.com") || containsString(body, "secret-key"))
}

func containsString(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
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
