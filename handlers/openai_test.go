package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"model-router/models"
	"model-router/services"
)

func newTestOpenAIHandler() http.HandlerFunc {
	registry := services.NewRegistry([]models.InternalModel{
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
	return NewOpenAIHandler(registry, forwarder)
}

func newTestOpenAIHandlerWithServer() (http.HandlerFunc, *httptest.Server) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"choices":[{"message":{"content":"test response"}}]}`))
	}))
	registry := services.NewRegistry([]models.InternalModel{
		{
			Name:          "test-model",
			RequestFormat: models.FormatOpenAI,
			Strategy:      models.StrategyFallback,
			RetryDelaySecs: 0,
			Externals: []models.ExternalModel{
				{Name: "test-external", URL: server.URL, Format: models.FormatOpenAI},
			},
		},
	})
	forwarder := services.NewForwarder()
	return NewOpenAIHandler(registry, forwarder), server
}

func TestOpenAIHandler_InvalidJSON(t *testing.T) {
	handler := newTestOpenAIHandler()

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader([]byte("{invalid}")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusBadRequest)
	}

	body, _ := io.ReadAll(rec.Body)
	var errResp map[string]interface{}
	json.Unmarshal(body, &errResp)

	if errResp["error"] == nil {
		t.Error("expected error field in response")
	}
}

func TestOpenAIHandler_EmptyModel(t *testing.T) {
	handler := newTestOpenAIHandler()

	body := `{"model": "", "messages": [{"role": "user", "content": "hello"}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader([]byte(body)))
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

func TestOpenAIHandler_EmptyMessages(t *testing.T) {
	handler := newTestOpenAIHandler()

	body := `{"model": "test-model", "messages": []}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader([]byte(body)))
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

func TestOpenAIHandler_ModelNotFound(t *testing.T) {
	handler := newTestOpenAIHandler()

	body := `{"model": "unknown-model", "messages": [{"role": "user", "content": "hello"}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader([]byte(body)))
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

func TestOpenAIHandler_NonStreaming_Success(t *testing.T) {
	handler, server := newTestOpenAIHandlerWithServer()
	defer server.Close()

	body := `{"model": "test-model", "messages": [{"role": "user", "content": "hello"}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusOK)
	}

	respBody, _ := io.ReadAll(rec.Body)
	var resp map[string]interface{}
	json.Unmarshal(respBody, &resp)
	if resp["choices"] == nil {
		t.Error("expected choices field in response")
	}
}

func TestOpenAIHandler_Stream_Success(t *testing.T) {
	sseServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`data: {"choices":[{"delta":{"content":"hello"}}]}\n\n`))
		w.(http.Flusher).Flush()
		w.Write([]byte(`data: [DONE]\n\n`))
		w.(http.Flusher).Flush()
	}))
	defer sseServer.Close()

	registry := services.NewRegistry([]models.InternalModel{
		{
			Name:          "test-model",
			RequestFormat: models.FormatOpenAI,
			Strategy:      models.StrategyFallback,
			RetryDelaySecs: 0,
			Externals: []models.ExternalModel{
				{Name: "test-external", URL: sseServer.URL, Format: models.FormatOpenAI},
			},
		},
	})
	forwarder := services.NewForwarder()
	handler := NewOpenAIHandler(registry, forwarder)

	body := `{"model": "test-model", "messages": [{"role": "user", "content": "hello"}], "stream": true}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusOK)
	}
	respBody := rec.Body.String()
	if !strings.Contains(respBody, `"content":"hello"`) {
		t.Errorf("response missing SSE chunk, got: %q", respBody)
	}
}

func TestOpenAIHandler_EmptyExternals(t *testing.T) {
	registry := services.NewRegistry([]models.InternalModel{
		{
			Name:          "empty-model",
			RequestFormat: models.FormatOpenAI,
			Strategy:      models.StrategyFallback,
			Externals:     []models.ExternalModel{},
		},
	})
	forwarder := services.NewForwarder()
	handler := NewOpenAIHandler(registry, forwarder)

	body := `{"model": "empty-model", "messages": [{"role": "user", "content": "hi"}], "stream": true}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestOpenAIHandler_Fallback_SecondProviderSucceeds(t *testing.T) {
	failServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(`{}`))
	}))
	okServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"choices":[{"message":{"content":"fallback succeeded"}}]}`))
	}))
	defer failServer.Close()
	defer okServer.Close()

	registry := services.NewRegistry([]models.InternalModel{
		{
			Name:          "test-model",
			RequestFormat: models.FormatOpenAI,
			Strategy:      models.StrategyFallback,
			RetryDelaySecs: 0,
			Externals: []models.ExternalModel{
				{Name: "fail", URL: failServer.URL, Format: models.FormatOpenAI},
				{Name: "ok", URL: okServer.URL, Format: models.FormatOpenAI},
			},
		},
	})
	forwarder := services.NewForwarder()
	handler := NewOpenAIHandler(registry, forwarder)

	body := `{"model": "test-model", "messages": [{"role": "user", "content": "hello"}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusOK)
	}
	respBody, _ := io.ReadAll(rec.Body)
	var resp map[string]interface{}
	json.Unmarshal(respBody, &resp)
	content := resp["choices"].([]interface{})[0].(map[string]interface{})["message"].(map[string]interface{})["content"]
	if content != "fallback succeeded" {
		t.Errorf("content = %q; want %q", content, "fallback succeeded")
	}
}

func TestOpenAIHandler_Fallback_AllFail(t *testing.T) {
	failServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(`{"error":"provider down"}`))
	}))
	defer failServer.Close()

	registry := services.NewRegistry([]models.InternalModel{
		{
			Name:          "test-model",
			RequestFormat: models.FormatOpenAI,
			Strategy:      models.StrategyFallback,
			RetryDelaySecs: 0,
			Externals: []models.ExternalModel{
				{Name: "fail-1", URL: failServer.URL, Format: models.FormatOpenAI},
				{Name: "fail-2", URL: failServer.URL, Format: models.FormatOpenAI},
			},
		},
	})
	forwarder := services.NewForwarder()
	handler := NewOpenAIHandler(registry, forwarder)

	body := `{"model": "test-model", "messages": [{"role": "user", "content": "hello"}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusBadGateway)
	}
}
