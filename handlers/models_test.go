package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"model-router/models"
	"model-router/services"
)

func TestModelsHandler_List(t *testing.T) {
	var registry services.RegistryReader = services.NewRegistry([]models.InternalModel{
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

	app := fiber.New()
	app.Get("/models", handler.List)

	req := httptest.NewRequest(http.MethodGet, "/models", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d; want %d", resp.StatusCode, http.StatusOK)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	models, ok := result["models"].([]interface{})
	if !ok {
		t.Fatal("expected models field to be an array")
	}
	if len(models) != 2 {
		t.Errorf("len(models) = %d; want 2", len(models))
	}
}

func TestModelsHandler_List_Empty(t *testing.T) {
	var registry services.RegistryReader = services.NewRegistry([]models.InternalModel{})
	handler := NewModelsHandler(registry)

	app := fiber.New()
	app.Get("/models", handler.List)

	req := httptest.NewRequest(http.MethodGet, "/models", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d; want %d", resp.StatusCode, http.StatusOK)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	models, ok := result["models"].([]interface{})
	if !ok {
		t.Fatal("expected models field to be an array")
	}
	if len(models) != 0 {
		t.Errorf("len(models) = %d; want 0", len(models))
	}
}