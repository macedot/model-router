package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReadBody_ValidJSON(t *testing.T) {
	body := `{"model":"test","messages":[{"role":"user","content":"hi"}],"temperature":0.7,"top_p":0.9}`
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	mapBody, envelope, err := readBody(req)
	if err != nil {
		t.Fatalf("readBody() error = %v", err)
	}

	if envelope.Model != "test" {
		t.Errorf("envelope.Model = %q, want %q", envelope.Model, "test")
	}
	if !envelope.HasMessages() {
		t.Error("expected HasMessages() = true")
	}
	if mapBody["temperature"] != 0.7 {
		t.Errorf("map temperature = %v, want 0.7", mapBody["temperature"])
	}
	if mapBody["top_p"] != 0.9 {
		t.Errorf("map top_p = %v, want 0.9", mapBody["top_p"])
	}
}

func TestReadBody_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`{invalid}`)))
	req.Header.Set("Content-Type", "application/json")

	_, _, err := readBody(req)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestReadBody_EmptyBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(``)))

	_, _, err := readBody(req)
	if err == nil {
		t.Fatal("expected error for empty body, got nil")
	}
}

func TestReadBody_NonObjectJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`"just a string"`)))

	_, _, err := readBody(req)
	if err == nil {
		t.Fatal("expected error for non-object JSON, got nil")
	}
}

func TestReadBody_StreamField(t *testing.T) {
	body := `{"model":"test","messages":[{"role":"user","content":"hi"}],"stream":true}`
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(body)))

	_, envelope, err := readBody(req)
	if err != nil {
		t.Fatalf("readBody() error = %v", err)
	}
	if envelope.Stream == nil {
		t.Fatal("expected Stream to be set")
	}
	if !*envelope.Stream {
		t.Error("expected Stream = true")
	}
}
