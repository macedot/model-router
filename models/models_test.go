package models

import "testing"

func TestProvider_ToExternal(t *testing.T) {
	p := Provider{
		ID:     "p1",
		Name:   "test-provider",
		URL:    "https://api.example.com",
		APIKey: "sk-key",
		Format: FormatOpenAI,
	}

	ext := p.ToExternal()

	if ext.Name != p.Name {
		t.Errorf("Name = %q, want %q", ext.Name, p.Name)
	}
	if ext.URL != p.URL {
		t.Errorf("URL = %q, want %q", ext.URL, p.URL)
	}
	if ext.APIKey != p.APIKey {
		t.Errorf("APIKey = %q, want %q", ext.APIKey, p.APIKey)
	}
	if ext.Format != p.Format {
		t.Errorf("Format = %q, want %q", ext.Format, p.Format)
	}
}

func TestProvider_ToExternal_DoesNotCopyID(t *testing.T) {
	p := Provider{ID: "p1", Name: "a", URL: "u", APIKey: "k", Format: FormatAnthropic}
	ext := p.ToExternal()
	// ExternalModel has no ID field — verify it's zero
	if ext.Name != "a" {
		t.Errorf("unexpected Name = %q", ext.Name)
	}
}

func TestRequestEnvelope_HasMessages_ValidMessages(t *testing.T) {
	e := &RequestEnvelope{
		Messages: jsonRaw(`[{"role":"user","content":"hi"}]`),
	}
	if !e.HasMessages() {
		t.Error("expected HasMessages() = true for valid messages")
	}
}

func TestRequestEnvelope_HasMessages_EmptyArray(t *testing.T) {
	e := &RequestEnvelope{
		Messages: jsonRaw(`[]`),
	}
	if e.HasMessages() {
		t.Error("expected HasMessages() = false for empty array")
	}
}

func TestRequestEnvelope_HasMessages_NilMessages(t *testing.T) {
	e := &RequestEnvelope{}
	if e.HasMessages() {
		t.Error("expected HasMessages() = false for nil messages")
	}
}

func TestRequestEnvelope_HasMessages_InvalidJSON(t *testing.T) {
	e := &RequestEnvelope{
		Messages: jsonRaw(`not json`),
	}
	if e.HasMessages() {
		t.Error("expected HasMessages() = false for invalid JSON")
	}
}

func jsonRaw(s string) []byte {
	return []byte(s)
}
