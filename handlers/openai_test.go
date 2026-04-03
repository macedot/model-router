package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"model-router/models"
	"model-router/services"
)

func newTestOpenAIHandler() (*OpenAIHandler, *fiber.App) {
	var registry services.RegistryReader = services.NewRegistry([]models.InternalModel{
		{
			Name:          "test-model",
			RequestFormat: models.FormatOpenAI,
			Strategy:      models.StrategyFallback,
			RetryDelaySecs: 0,
			Externals: []models.ExternalModel{
				{Name: "test-external", URL: "https://api.example.com", Format: models.FormatOpenAI},
			},
		},
	})
	forwarder := services.NewForwarder()
	handler := NewOpenAIHandler(registry, forwarder)

	app := fiber.New()
	app.Post("/v1/chat/completions", handler.Handle)
	return handler, app
}

func TestOpenAIHandler_InvalidJSON(t *testing.T) {
	_, app := newTestOpenAIHandler()

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader([]byte("{invalid}")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("status = %d; want %d", resp.StatusCode, fiber.StatusBadRequest)
	}

	body, _ := io.ReadAll(resp.Body)
	var errResp map[string]interface{}
	json.Unmarshal(body, &errResp)

	if errResp["error"] == nil {
		t.Error("expected error field in response")
	}
}

func TestOpenAIHandler_EmptyModel(t *testing.T) {
	_, app := newTestOpenAIHandler()

	body := `{"model": "", "messages": [{"role": "user", "content": "hello"}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("status = %d; want %d", resp.StatusCode, fiber.StatusBadRequest)
	}

	respBody, _ := io.ReadAll(resp.Body)
	var errResp map[string]interface{}
	json.Unmarshal(respBody, &errResp)

	errMap := errResp["error"].(map[string]interface{})
	if errMap["message"] != "model is required" {
		t.Errorf("message = %q; want %q", errMap["message"], "model is required")
	}
}

func TestOpenAIHandler_EmptyMessages(t *testing.T) {
	_, app := newTestOpenAIHandler()

	body := `{"model": "test-model", "messages": []}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("status = %d; want %d", resp.StatusCode, fiber.StatusBadRequest)
	}

	respBody, _ := io.ReadAll(resp.Body)
	var errResp map[string]interface{}
	json.Unmarshal(respBody, &errResp)

	errMap := errResp["error"].(map[string]interface{})
	if errMap["message"] != "messages is required and cannot be empty" {
		t.Errorf("message = %q; want %q", errMap["message"], "messages is required and cannot be empty")
	}
}

func TestOpenAIHandler_ModelNotFound(t *testing.T) {
	_, app := newTestOpenAIHandler()

	body := `{"model": "unknown-model", "messages": [{"role": "user", "content": "hello"}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusNotFound {
		t.Errorf("status = %d; want %d", resp.StatusCode, fiber.StatusNotFound)
	}

	respBody, _ := io.ReadAll(resp.Body)
	var errResp map[string]interface{}
	json.Unmarshal(respBody, &errResp)

	errMap := errResp["error"].(map[string]interface{})
	if errMap["message"] != "Model not found: unknown-model" {
		t.Errorf("message = %q; want %q", errMap["message"], "Model not found: unknown-model")
	}
}