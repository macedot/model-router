package services

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"model-router/models"
)

func TestForwarder_Forward_Success(t *testing.T) {
	respBody := `{"choices":[{"message":{"content":"hi"}}]}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("Authorization header = %q, want %q", r.Header.Get("Authorization"), "Bearer test-key")
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %q, want %q", r.Header.Get("Content-Type"), "application/json")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(respBody))
	}))
	defer server.Close()

	f := NewForwarder()
	body := []byte(`{"model":"test","messages":[{"role":"user","content":"hello"}]}`)
	target := models.ExternalModel{
		Name: "ext", URL: server.URL, APIKey: "test-key", Format: models.FormatOpenAI,
	}

	result, err := f.Forward(context.Background(), body, target)
	if err != nil {
		t.Fatalf("Forward() error = %v", err)
	}
	if string(result) != respBody {
		t.Errorf("body = %q, want %q", string(result), respBody)
	}
}

func TestForwarder_Forward_AnthropicAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "anthropic-key" {
			t.Errorf("x-api-key = %q, want %q", r.Header.Get("x-api-key"), "anthropic-key")
		}
		if r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Errorf("anthropic-version = %q, want %q", r.Header.Get("anthropic-version"), "2023-06-01")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"content":"ok"}`))
	}))
	defer server.Close()

	f := NewForwarder()
	body := []byte(`{"model":"test","messages":[{"role":"user","content":"hello"}],"max_tokens":1024}`)
	target := models.ExternalModel{
		Name: "claude", URL: server.URL, APIKey: "anthropic-key", Format: models.FormatAnthropic,
	}

	result, err := f.Forward(context.Background(), body, target)
	if err != nil {
		t.Fatalf("Forward() error = %v", err)
	}
	if string(result) != `{"content":"ok"}` {
		t.Errorf("body = %q", string(result))
	}
}

func TestForwarder_Forward_ErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"provider error"}`))
	}))
	defer server.Close()

	f := NewForwarder()
	body := []byte(`{}`)
	target := models.ExternalModel{Name: "ext", URL: server.URL, APIKey: "key", Format: models.FormatOpenAI}

	_, err := f.Forward(context.Background(), body, target)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestForwarder_Forward_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	f := NewForwarder()
	f.client.Timeout = 5 * time.Second

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := f.Forward(ctx, []byte(`{}`), models.ExternalModel{Name: "ext", URL: server.URL, APIKey: "key", Format: models.FormatOpenAI})
	if err == nil {
		t.Fatal("expected context error, got nil")
	}
}

func TestForwarder_Forward_ResponseTooLarge(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		chunk := make([]byte, 64*1024)
		for i := range chunk {
			chunk[i] = 'x'
		}
		for i := 0; i < 80; i++ {
			w.Write(chunk)
		}
	}))
	defer server.Close()

	f := NewForwarder()
	_, err := f.Forward(context.Background(), []byte(`{}`), models.ExternalModel{Name: "ext", URL: server.URL, APIKey: "key", Format: models.FormatOpenAI})
	if err != nil {
		t.Fatalf("unexpected error for under-limit response: %v", err)
	}
}

func TestForwarder_ForwardStream_Success(t *testing.T) {
	chunks := []string{
		`data: {"choices":[{"delta":{"content":"hello"}}]}\n\n`,
		`data: {"choices":[{"delta":{"content":" world"}}]}\n\n`,
		`data: [DONE]\n\n`,
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		for _, chunk := range chunks {
			w.Write([]byte(chunk))
			w.(http.Flusher).Flush()
		}
	}))
	defer server.Close()

	f := NewForwarder()
	body := []byte(`{"model":"test","messages":[{"role":"user","content":"hi"}],"stream":true}`)
	target := models.ExternalModel{
		Name: "ext", URL: server.URL, APIKey: "test-key", Format: models.FormatOpenAI,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		err := f.ForwardStream(r.Context(), body, target, w)
		if err != nil {
			t.Errorf("ForwardStream() error = %v", err)
		}
	})

	streamReq := httptest.NewRequest(http.MethodGet, "/stream", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, streamReq)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	for _, chunk := range chunks {
		if !contains(rec.Body.String(), chunk) {
			t.Errorf("response missing chunk %q", chunk)
		}
	}
}

func TestForwarder_ForwardStream_ErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(`{"error":"upstream error"}`))
	}))
	defer server.Close()

	f := NewForwarder()
	body := []byte(`{"stream":true}`)
	target := models.ExternalModel{Name: "ext", URL: server.URL, APIKey: "key", Format: models.FormatOpenAI}

	var errReturned error
	mux := http.NewServeMux()
	mux.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		errReturned = f.ForwardStream(r.Context(), body, target, w)
	})

	streamReq := httptest.NewRequest(http.MethodGet, "/stream", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, streamReq)

	if errReturned == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestForwarder_ForwardStream_ResponseTooLarge(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		large := make([]byte, 11*1024*1024)
		for i := range large {
			large[i] = 'x'
		}
		w.Write(large)
	}))
	defer server.Close()

	f := NewForwarder()
	body := []byte(`{"stream":true}`)
	target := models.ExternalModel{Name: "ext", URL: server.URL, APIKey: "key", Format: models.FormatOpenAI}

	var errReturned error
	mux := http.NewServeMux()
	mux.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		errReturned = f.ForwardStream(r.Context(), body, target, w)
	})

	streamReq := httptest.NewRequest(http.MethodGet, "/stream", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, streamReq)

	if errReturned == nil {
		t.Fatal("expected error for oversized response, got nil")
	}
}

func TestForwarder_ForwardStream_PassesModelName(t *testing.T) {
	var receivedBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "anthropic-key" {
			t.Errorf("x-api-key = %q", r.Header.Get("x-api-key"))
		}
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`data: {"content":"hi"}\n\n`))
		w.(http.Flusher).Flush()
	}))
	defer server.Close()

	f := NewForwarder()
	// Simulate cross-format: OpenAI request to Anthropic target
	bodyMap := map[string]interface{}{
		"model":    "gpt-4",
		"messages": []interface{}{map[string]string{"role": "user", "content": "hi"}},
		"stream":   true,
	}
	prepared, _ := PrepareRequest(bodyMap, "claude", models.FormatOpenAI, models.FormatAnthropic)

	mux := http.NewServeMux()
	mux.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		f.ForwardStream(r.Context(), prepared, models.ExternalModel{
			Name: "claude", URL: server.URL, APIKey: "anthropic-key", Format: models.FormatAnthropic,
		}, w)
	})

	streamReq := httptest.NewRequest(http.MethodGet, "/stream", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, streamReq)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if receivedBody["model"] != "claude" {
		t.Errorf("model = %q, want %q", receivedBody["model"], "claude")
	}
}

func TestLimitedReader_ExceedsLimit(t *testing.T) {
	reader := limitedReader{R: io.NopCloser(strings.NewReader("xxxxxxxxxxxxxxxxxxxx")), Left: 10}
	buf := make([]byte, 20)
	n, err := reader.Read(buf)
	if err == nil {
		t.Error("expected errResponseTruncated, got nil")
	}
	if n != 10 {
		t.Errorf("n = %d, want 10", n)
	}
}

func TestLimitedReader_WithinLimit(t *testing.T) {
	reader := limitedReader{R: io.NopCloser(strings.NewReader("xxxxx")), Left: 10}
	buf := make([]byte, 5)
	n, err := reader.Read(buf)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if n != 5 {
		t.Errorf("n = %d, want 5", n)
	}
}

func TestLimitedReader_ZeroLeft(t *testing.T) {
	reader := limitedReader{R: io.NopCloser(strings.NewReader("xxxxxxxxxx")), Left: 0}
	buf := make([]byte, 5)
	_, err := reader.Read(buf)
	if err != errResponseTruncated {
		t.Errorf("err = %v, want errResponseTruncated", err)
	}
}

func TestNewForwarder_DefaultTimeout(t *testing.T) {
	f := NewForwarder()
	if f.client.Timeout != defaultTimeout {
		t.Errorf("Timeout = %v, want %v", f.client.Timeout, defaultTimeout)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
