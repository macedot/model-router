package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"model-router/models"
	"model-router/services"
)

func newTestAnthropicHandler() http.HandlerFunc {
	registry := services.NewRegistry([]models.InternalModel{
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
	return NewAnthropicHandler(registry, forwarder)
}

func TestAnthropicHandler_InvalidJSON(t *testing.T) {
	handler := newTestAnthropicHandler()

	req := httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader([]byte("{invalid}")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestAnthropicHandler_EmptyModel(t *testing.T) {
	handler := newTestAnthropicHandler()

	body := `{"model": "", "messages": [{"role": "user", "content": "hello"}], "max_tokens": 1024}`
	req := httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusBadRequest)
	}

	respBody, _ := io.ReadAll(rec.Body)
	var errResp map[string]interface{}
	json.Unmarshal(respBody, &errResp)

	errMap := errResp["error"].(map[string]interface{})
	if errMap["message"] != "model is required" {
		t.Errorf("message = %q; want %q", errMap["message"], "model is required")
	}
}

func TestAnthropicHandler_EmptyMessages(t *testing.T) {
	handler := newTestAnthropicHandler()

	body := `{"model": "test-anthropic-model", "messages": [], "max_tokens": 1024}`
	req := httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusBadRequest)
	}

	respBody, _ := io.ReadAll(rec.Body)
	var errResp map[string]interface{}
	json.Unmarshal(respBody, &errResp)

	errMap := errResp["error"].(map[string]interface{})
	if errMap["message"] != "messages is required and cannot be empty" {
		t.Errorf("message = %q; want %q", errMap["message"], "messages is required and cannot be empty")
	}
}

func TestAnthropicHandler_ModelNotFound(t *testing.T) {
	handler := newTestAnthropicHandler()

	body := `{"model": "unknown-model", "messages": [{"role": "user", "content": "hello"}], "max_tokens": 1024}`
	req := httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusNotFound)
	}

	respBody, _ := io.ReadAll(rec.Body)
	var errResp map[string]interface{}
	json.Unmarshal(respBody, &errResp)

	errMap := errResp["error"].(map[string]interface{})
	if errMap["message"] != "Model not found: unknown-model" {
		t.Errorf("message = %q; want %q", errMap["message"], "Model not found: unknown-model")
	}
}

func TestAnthropicHandler_Success(t *testing.T) {
	respBody := `{"id":"msg_123","type":"message","content":[{"type":"text","text":"hello"}]}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(respBody))
	}))
	defer server.Close()

	registry := services.NewRegistry([]models.InternalModel{
		{
			Name:          "test-model",
			RequestFormat: models.FormatAnthropic,
			Externals: []models.ExternalModel{
				{Name: "ext", URL: server.URL, APIKey: "key", Format: models.FormatAnthropic},
			},
		},
	})
	handler := NewAnthropicHandler(registry, services.NewForwarder())

	body := `{"model": "test-model", "messages": [{"role": "user", "content": "hi"}], "max_tokens": 1024}`
	req := httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != respBody {
		t.Errorf("body = %q, want %q", rec.Body.String(), respBody)
	}
}

func TestAnthropicHandler_Fallback(t *testing.T) {
	failServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"fail"}`))
	}))
	defer failServer.Close()

	okServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"content":"ok"}`))
	}))
	defer okServer.Close()

	registry := services.NewRegistry([]models.InternalModel{
		{
			Name:          "test-model",
			RequestFormat: models.FormatAnthropic,
			Strategy:      models.StrategyFallback,
			Externals: []models.ExternalModel{
				{Name: "fail", URL: failServer.URL, APIKey: "key", Format: models.FormatAnthropic},
				{Name: "ok", URL: okServer.URL, APIKey: "key", Format: models.FormatAnthropic},
			},
		},
	})
	handler := NewAnthropicHandler(registry, services.NewForwarder())

	body := `{"model": "test-model", "messages": [{"role": "user", "content": "hi"}], "max_tokens": 1024}`
	req := httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != `{"content":"ok"}` {
		t.Errorf("body = %q", rec.Body.String())
	}
}

func TestAnthropicHandler_AllProvidersFail(t *testing.T) {
	failServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"fail"}`))
	}))
	defer failServer.Close()

	registry := services.NewRegistry([]models.InternalModel{
		{
			Name:          "test-model",
			RequestFormat: models.FormatAnthropic,
			Strategy:      models.StrategyFallback,
			Externals: []models.ExternalModel{
				{Name: "fail-1", URL: failServer.URL, APIKey: "key", Format: models.FormatAnthropic},
				{Name: "fail-2", URL: failServer.URL, APIKey: "key", Format: models.FormatAnthropic},
			},
		},
	})
	handler := NewAnthropicHandler(registry, services.NewForwarder())

	body := `{"model": "test-model", "messages": [{"role": "user", "content": "hi"}], "max_tokens": 1024}`
	req := httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusBadGateway)
	}
}
