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

func newTestAnthropicHandler() (*AnthropicHandler, *fiber.App) {
	var registry services.RegistryReader = services.NewRegistry([]models.InternalModel{
		{
			Name:          "test-anthropic-model",
			RequestFormat: models.FormatAnthropic,
			Strategy:      models.StrategyFallback,
			RetryDelaySecs: 0,
			Externals: []models.ExternalModel{
				{Name: "anthropic-external", URL: "https://api.anthropic.com", Format: models.FormatAnthropic},
			},
		},
	})
	forwarder := services.NewForwarder()
	handler := NewAnthropicHandler(registry, forwarder)

	app := fiber.New()
	app.Post("/v1/messages", handler.Handle)
	return handler, app
}

func TestAnthropicHandler_InvalidJSON(t *testing.T) {
	_, app := newTestAnthropicHandler()

	req := httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader([]byte("{invalid}")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("status = %d; want %d", resp.StatusCode, fiber.StatusBadRequest)
	}
}

func TestAnthropicHandler_EmptyModel(t *testing.T) {
	_, app := newTestAnthropicHandler()

	body := `{"model": "", "messages": [{"role": "user", "content": "hello"}], "max_tokens": 1024}`
	req := httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader([]byte(body)))
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

func TestAnthropicHandler_EmptyMessages(t *testing.T) {
	_, app := newTestAnthropicHandler()

	body := `{"model": "test-anthropic-model", "messages": [], "max_tokens": 1024}`
	req := httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader([]byte(body)))
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

func TestAnthropicHandler_ModelNotFound(t *testing.T) {
	_, app := newTestAnthropicHandler()

	body := `{"model": "unknown-model", "messages": [{"role": "user", "content": "hello"}], "max_tokens": 1024}`
	req := httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader([]byte(body)))
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