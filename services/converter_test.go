package services

import (
	"encoding/json"
	"testing"

	"model-router/models"
)

func parseJSON(t *testing.T, s string) map[string]interface{} {
	t.Helper()
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestPrepareRequest_SameFormat_Passthrough(t *testing.T) {
	body := parseJSON(t, `{
		"model": "test-model",
		"messages": [{"role": "user", "content": "hello"}],
		"temperature": 0.7,
		"top_p": 0.9,
		"max_tokens": 1024,
		"stream": true,
		"custom_field": "preserved"
	}`)

	result, err := PrepareRequest(body, "target-model", models.FormatOpenAI, models.FormatOpenAI)
	if err != nil {
		t.Fatalf("PrepareRequest() error = %v", err)
	}

	var parsed map[string]interface{}
	json.Unmarshal(result, &parsed)

	if parsed["model"] != "target-model" {
		t.Errorf("model = %q, want %q", parsed["model"], "target-model")
	}
	if parsed["temperature"] != 0.7 {
		t.Errorf("temperature = %v, want 0.7", parsed["temperature"])
	}
	if parsed["top_p"] != 0.9 {
		t.Errorf("top_p = %v, want 0.9", parsed["top_p"])
	}
	if parsed["max_tokens"] != 1024.0 {
		t.Errorf("max_tokens = %v, want 1024", parsed["max_tokens"])
	}
	if parsed["custom_field"] != "preserved" {
		t.Errorf("custom_field = %q, want %q", parsed["custom_field"], "preserved")
	}
}

func TestPrepareRequest_CrossFormat_FieldRenames(t *testing.T) {
	body := parseJSON(t, `{
		"model": "test",
		"messages": [{"role": "user", "content": "hi"}],
		"stop": ["end"],
		"max_completion_tokens": 2048,
		"temperature": 0.5
	}`)

	result, err := PrepareRequest(body, "claude-3", models.FormatOpenAI, models.FormatAnthropic)
	if err != nil {
		t.Fatalf("PrepareRequest() error = %v", err)
	}

	var parsed map[string]interface{}
	json.Unmarshal(result, &parsed)

	// stop → stop_sequences
	if parsed["stop_sequences"] == nil {
		t.Error("expected stop_sequences to be set")
	}
	if parsed["stop"] != nil {
		t.Error("expected stop to be removed")
	}
	// max_completion_tokens → max_tokens
	if parsed["max_tokens"] != 2048.0 {
		t.Errorf("max_tokens = %v, want 2048", parsed["max_tokens"])
	}
	if parsed["max_completion_tokens"] != nil {
		t.Error("expected max_completion_tokens to be removed")
	}
	// model renamed
	if parsed["model"] != "claude-3" {
		t.Errorf("model = %q, want %q", parsed["model"], "claude-3")
	}
	// passthrough field
	if parsed["temperature"] != 0.5 {
		t.Errorf("temperature = %v, want 0.5", parsed["temperature"])
	}
}

func TestPrepareRequest_CrossFormat_AnthropicToOpenAI(t *testing.T) {
	body := parseJSON(t, `{
		"model": "claude-3",
		"messages": [{"role": "user", "content": "hi"}],
		"max_tokens": 4096,
		"stop_sequences": ["END"],
		"temperature": 0.8
	}`)

	result, err := PrepareRequest(body, "gpt-4", models.FormatAnthropic, models.FormatOpenAI)
	if err != nil {
		t.Fatalf("PrepareRequest() error = %v", err)
	}

	var parsed map[string]interface{}
	json.Unmarshal(result, &parsed)

	// stop_sequences → stop
	if parsed["stop"] == nil {
		t.Error("expected stop to be set")
	}
	if parsed["stop_sequences"] != nil {
		t.Error("expected stop_sequences to be removed")
	}
	if parsed["model"] != "gpt-4" {
		t.Errorf("model = %q, want %q", parsed["model"], "gpt-4")
	}
}

func TestPrepareRequest_CrossFormat_SourceOnlyFieldsDropped(t *testing.T) {
	body := parseJSON(t, `{
		"model": "test",
		"messages": [{"role": "user", "content": "hi"}],
		"frequency_penalty": 0.5,
		"presence_penalty": 0.3,
		"temperature": 0.7
	}`)

	result, err := PrepareRequest(body, "claude-3", models.FormatOpenAI, models.FormatAnthropic)
	if err != nil {
		t.Fatalf("PrepareRequest() error = %v", err)
	}

	var parsed map[string]interface{}
	json.Unmarshal(result, &parsed)

	if parsed["frequency_penalty"] != nil {
		t.Error("expected frequency_penalty to be dropped")
	}
	if parsed["presence_penalty"] != nil {
		t.Error("expected presence_penalty to be dropped")
	}
	if parsed["temperature"] != 0.7 {
		t.Errorf("temperature = %v, want 0.7 (passthrough field)", parsed["temperature"])
	}
}

func TestPrepareRequest_CrossFormat_AnthropicOnlyFieldsDropped(t *testing.T) {
	body := parseJSON(t, `{
		"model": "claude-3",
		"messages": [{"role": "user", "content": "hi"}],
		"max_tokens": 1024,
		"top_k": 50,
		"temperature": 0.7
	}`)

	result, err := PrepareRequest(body, "gpt-4", models.FormatAnthropic, models.FormatOpenAI)
	if err != nil {
		t.Fatalf("PrepareRequest() error = %v", err)
	}

	var parsed map[string]interface{}
	json.Unmarshal(result, &parsed)

	if parsed["top_k"] != nil {
		t.Error("expected top_k to be dropped")
	}
	if parsed["temperature"] != 0.7 {
		t.Errorf("temperature = %v, want 0.7 (passthrough field)", parsed["temperature"])
	}
}
