package services

import (
	"math"
	"testing"

	"model-router/models"
)

func TestToAnthropic(t *testing.T) {
	temp := 0.7
	maxTokens := 100
	req := &models.OpenAIRequest{
		Model:       "gpt-4",
		Messages:    []models.Message{{Role: "user", Content: "hello"}},
		Temperature: &temp,
		MaxTokens:   &maxTokens,
	}

	anthropicReq := ToAnthropic(req)

	if anthropicReq.Model != "gpt-4" {
		t.Errorf("Model = %q, want %q", anthropicReq.Model, "gpt-4")
	}
	if len(anthropicReq.Messages) != 1 {
		t.Fatalf("Messages len = %d, want 1", len(anthropicReq.Messages))
	}
	if anthropicReq.Messages[0].Role != "user" {
		t.Errorf("Messages[0].Role = %q, want %q", anthropicReq.Messages[0].Role, "user")
	}
	if anthropicReq.Messages[0].Content != "hello" {
		t.Errorf("Messages[0].Content = %q, want %q", anthropicReq.Messages[0].Content, "hello")
	}
	if anthropicReq.MaxTokens != 100 {
		t.Errorf("MaxTokens = %d, want %d", anthropicReq.MaxTokens, 100)
	}
	if anthropicReq.Temperature == nil || *anthropicReq.Temperature != 0.7 {
		t.Errorf("Temperature = %v, want 0.7", anthropicReq.Temperature)
	}
}

func TestToAnthropic_DefaultMaxTokens(t *testing.T) {
	req := &models.OpenAIRequest{
		Model:    "gpt-4",
		Messages: []models.Message{{Role: "user", Content: "hello"}},
	}

	anthropicReq := ToAnthropic(req)

	if anthropicReq.MaxTokens != defaultMaxTokens {
		t.Errorf("MaxTokens = %d, want %d (default)", anthropicReq.MaxTokens, defaultMaxTokens)
	}
}

func TestToAnthropic_NilOptionalFields(t *testing.T) {
	req := &models.OpenAIRequest{
		Model:    "gpt-4",
		Messages: []models.Message{{Role: "user", Content: "hello"}},
	}

	anthropicReq := ToAnthropic(req)

	if anthropicReq.Temperature != nil {
		t.Errorf("Temperature = %v, want nil", anthropicReq.Temperature)
	}
}

func TestToAnthropic_Immutability(t *testing.T) {
	req := &models.OpenAIRequest{
		Model:       "gpt-4",
		Messages:    []models.Message{{Role: "user", Content: "hello"}},
		Temperature: floatPtr(0.7),
		MaxTokens:   intPtr(100),
	}

	originalModel := req.Model
	anthropicReq := ToAnthropic(req)

	// Verify original request is unchanged
	if req.Model != originalModel {
		t.Errorf("Original request Model was modified")
	}
	if req.Temperature == nil || *req.Temperature != 0.7 {
		t.Errorf("Original request Temperature was modified")
	}

	// Verify the converted request has the correct model
	if anthropicReq.Model != originalModel {
		t.Errorf("Converted request Model = %q, want %q", anthropicReq.Model, originalModel)
	}
}

func floatPtr(f float64) *float64 {
	return &f
}

func intPtr(i int) *int {
	return &i
}

func TestToAnthropic_NanHandling(t *testing.T) {
	nan := math.NaN()
	req := &models.OpenAIRequest{
		Model:       "gpt-4",
		Messages:    []models.Message{{Role: "user", Content: "hello"}},
		Temperature: &nan,
	}

	anthropicReq := ToAnthropic(req)

	// NaN should pass through (we don't validate)
	if anthropicReq.Temperature == nil || !math.IsNaN(*anthropicReq.Temperature) {
		t.Errorf("Temperature NaN not preserved")
	}
}