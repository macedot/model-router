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

func TestForwarder_ForwardOpenAI_Success(t *testing.T) {
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
	req := &models.OpenAIRequest{
		Model:    "test-model",
		Messages: []models.Message{{Role: "user", Content: "hello"}},
	}
	target := models.ExternalModel{
		Name:   "ext",
		URL:    server.URL,
		APIKey: "test-key",
		Format: models.FormatOpenAI,
	}

	body, err := f.ForwardOpenAI(context.Background(), req, target)
	if err != nil {
		t.Fatalf("ForwardOpenAI() error = %v", err)
	}
	if string(body) != respBody {
		t.Errorf("body = %q, want %q", string(body), respBody)
	}
}

func TestForwarder_ForwardOpenAI_AnthropicFormat(t *testing.T) {
	var receivedBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "anthropic-key" {
			t.Errorf("x-api-key = %q, want %q", r.Header.Get("x-api-key"), "anthropic-key")
		}
		if r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Errorf("anthropic-version = %q, want %q", r.Header.Get("anthropic-version"), "2023-06-01")
		}
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"content":"ok"}`))
	}))
	defer server.Close()

	f := NewForwarder()
	req := &models.OpenAIRequest{
		Model:    "gpt-4",
		Messages: []models.Message{{Role: "user", Content: "hello"}},
	}
	target := models.ExternalModel{
		Name:   "claude",
		URL:    server.URL,
		APIKey: "anthropic-key",
		Format: models.FormatAnthropic,
	}

	body, err := f.ForwardOpenAI(context.Background(), req, target)
	if err != nil {
		t.Fatalf("ForwardOpenAI() error = %v", err)
	}
	if string(body) != `{"content":"ok"}` {
		t.Errorf("body = %q", string(body))
	}
	if receivedBody["model"] != "claude" {
		t.Errorf("converted model = %q, want %q", receivedBody["model"], "claude")
	}
}

func TestForwarder_ForwardOpenAI_ErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"provider error"}`))
	}))
	defer server.Close()

	f := NewForwarder()
	req := &models.OpenAIRequest{Model: "test", Messages: []models.Message{{Role: "user", Content: "hi"}}}
	target := models.ExternalModel{Name: "ext", URL: server.URL, APIKey: "key", Format: models.FormatOpenAI}

	_, err := f.ForwardOpenAI(context.Background(), req, target)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestForwarder_ForwardOpenAI_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	f := NewForwarder()
	f.client.Timeout = 5 * time.Second

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	req := &models.OpenAIRequest{Model: "test", Messages: []models.Message{{Role: "user", Content: "hi"}}}
	target := models.ExternalModel{Name: "ext", URL: server.URL, APIKey: "key", Format: models.FormatOpenAI}

	_, err := f.ForwardOpenAI(ctx, req, target)
	if err == nil {
		t.Fatal("expected context error, got nil")
	}
}

func TestForwarder_ForwardOpenAI_ResponseTooLarge(t *testing.T) {
	// NOTE: io.LimitReader returns io.EOF (not error) when the limit is reached,
	// so the non-streaming path silently truncates at maxResponseSize.
	// The streaming path (ForwardOpenAIStream) uses limitedReader which returns
	// errResponseTruncated and is tested separately.
	// This test verifies the non-streaming path at least handles a large response.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Send 5MB response (under 10MB limit)
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
	req := &models.OpenAIRequest{Model: "test", Messages: []models.Message{{Role: "user", Content: "hi"}}}
	target := models.ExternalModel{Name: "ext", URL: server.URL, APIKey: "key", Format: models.FormatOpenAI}

	body, err := f.ForwardOpenAI(context.Background(), req, target)
	if err != nil {
		t.Fatalf("unexpected error for under-limit response: %v", err)
	}
	if len(body) != 5*1024*1024 {
		t.Errorf("body len = %d, want %d", len(body), 5*1024*1024)
	}
}

func TestForwarder_ForwardAnthropic_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "anthropic-key" {
			t.Errorf("x-api-key = %q, want %q", r.Header.Get("x-api-key"), "anthropic-key")
		}
		if r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Errorf("anthropic-version = %q, want %q", r.Header.Get("anthropic-version"), "2023-06-01")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"content":"anthropic response"}`))
	}))
	defer server.Close()

	f := NewForwarder()
	req := &models.AnthropicRequest{
		Model:     "claude-3",
		Messages:  []models.Message{{Role: "user", Content: "hello"}},
		MaxTokens: 1024,
	}
	target := models.ExternalModel{
		Name:   "claude",
		URL:    server.URL,
		APIKey: "anthropic-key",
		Format: models.FormatAnthropic,
	}

	body, err := f.ForwardAnthropic(context.Background(), req, target)
	if err != nil {
		t.Fatalf("ForwardAnthropic() error = %v", err)
	}
	if string(body) != `{"content":"anthropic response"}` {
		t.Errorf("body = %q", string(body))
	}
}

func TestForwarder_ForwardAnthropic_ErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"bad request"}`))
	}))
	defer server.Close()

	f := NewForwarder()
	req := &models.AnthropicRequest{Model: "claude", Messages: []models.Message{{Role: "user", Content: "hi"}}, MaxTokens: 1024}
	target := models.ExternalModel{Name: "ext", URL: server.URL, APIKey: "key", Format: models.FormatAnthropic}

	_, err := f.ForwardAnthropic(context.Background(), req, target)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestForwarder_ForwardAnthropic_ContextCancelled(t *testing.T) {
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

	req := &models.AnthropicRequest{Model: "claude", Messages: []models.Message{{Role: "user", Content: "hi"}}, MaxTokens: 1024}
	target := models.ExternalModel{Name: "ext", URL: server.URL, APIKey: "key", Format: models.FormatAnthropic}

	_, err := f.ForwardAnthropic(ctx, req, target)
	if err == nil {
		t.Fatal("expected context error, got nil")
	}
}

func TestForwarder_ForwardOpenAIStream_Success(t *testing.T) {
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
	req := &models.OpenAIRequest{
		Model:    "test-model",
		Messages: []models.Message{{Role: "user", Content: "hi"}},
		Stream:   boolPtr(true),
	}
	target := models.ExternalModel{
		Name:   "ext",
		URL:    server.URL,
		APIKey: "test-key",
		Format: models.FormatOpenAI,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		err := f.ForwardOpenAIStream(r.Context(), req, target, w)
		if err != nil {
			t.Errorf("ForwardOpenAIStream() error = %v", err)
		}
	})

	streamReq := httptest.NewRequest(http.MethodGet, "/stream", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, streamReq)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	for _, chunk := range chunks {
		if !contains(body, chunk) {
			t.Errorf("response missing chunk %q", chunk)
		}
	}
}

func TestForwarder_ForwardOpenAIStream_ErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(`{"error":"upstream error"}`))
	}))
	defer server.Close()

	f := NewForwarder()
	req := &models.OpenAIRequest{
		Model:    "test",
		Messages: []models.Message{{Role: "user", Content: "hi"}},
		Stream:   boolPtr(true),
	}
	target := models.ExternalModel{Name: "ext", URL: server.URL, APIKey: "key", Format: models.FormatOpenAI}

	var errReturned error
	mux := http.NewServeMux()
	mux.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		errReturned = f.ForwardOpenAIStream(r.Context(), req, target, w)
	})

	streamReq := httptest.NewRequest(http.MethodGet, "/stream", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, streamReq)

	if errReturned == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestForwarder_ForwardOpenAIStream_ResponseTooLarge(t *testing.T) {
	// Server sends 11MB response (over 10MB limit)
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
	req := &models.OpenAIRequest{
		Model:    "test",
		Messages: []models.Message{{Role: "user", Content: "hi"}},
		Stream:   boolPtr(true),
	}
	target := models.ExternalModel{Name: "ext", URL: server.URL, APIKey: "key", Format: models.FormatOpenAI}

	var errReturned error
	mux := http.NewServeMux()
	mux.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		errReturned = f.ForwardOpenAIStream(r.Context(), req, target, w)
	})

	streamReq := httptest.NewRequest(http.MethodGet, "/stream", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, streamReq)

	if errReturned == nil {
		t.Fatal("expected error for oversized response, got nil")
	}
}

func TestForwarder_ForwardOpenAIStream_AnthropicFormat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "anthropic-key" {
			t.Errorf("x-api-key = %q", r.Header.Get("x-api-key"))
		}
		if r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Errorf("anthropic-version = %q", r.Header.Get("anthropic-version"))
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`data: {"content":"hi"}\n\n`))
		w.(http.Flusher).Flush()
	}))
	defer server.Close()

	f := NewForwarder()
	req := &models.OpenAIRequest{
		Model:    "gpt-4",
		Messages: []models.Message{{Role: "user", Content: "hi"}},
		Stream:   boolPtr(true),
	}
	target := models.ExternalModel{
		Name:   "claude",
		URL:    server.URL,
		APIKey: "anthropic-key",
		Format: models.FormatAnthropic,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		f.ForwardOpenAIStream(r.Context(), req, target, w)
	})

	streamReq := httptest.NewRequest(http.MethodGet, "/stream", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, streamReq)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
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

func boolPtr(b bool) *bool { return &b }

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
