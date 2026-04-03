package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"model-router/models"
)

const (
	maxBodySize      = 10 * 1024 * 1024 // 10MB
	maxResponseSize  = 10 * 1024 * 1024 // 10MB
	defaultTimeout   = 120 * time.Second
	minSuccessStatus = 200
	maxSuccessStatus = 300
	anthropicVersion = "2023-06-01"
)

type Forwarder struct {
	client *http.Client
}

func NewForwarder() *Forwarder {
	return &Forwarder{
		client: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

func (f *Forwarder) ForwardOpenAI(ctx context.Context, req *models.OpenAIRequest, target models.ExternalModel) ([]byte, error) {
	var body []byte
	var err error

	if target.Format == models.FormatAnthropic {
		anthropicReq := ToAnthropic(req)
		anthropicReq.Model = target.Name
		body, err = json.Marshal(anthropicReq)
	} else {
		req.Model = target.Name
		body, err = json.Marshal(req)
	}
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, target.URL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if target.Format == models.FormatAnthropic {
		httpReq.Header.Set("x-api-key", target.APIKey)
		httpReq.Header.Set("anthropic-version", anthropicVersion)
	} else {
		httpReq.Header.Set("Authorization", "Bearer "+target.APIKey)
	}

	resp, err := f.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("forwarding request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode < minSuccessStatus || resp.StatusCode >= maxSuccessStatus {
		return nil, fmt.Errorf("external API returned %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func (f *Forwarder) ForwardOpenAIStream(ctx context.Context, req *models.OpenAIRequest, target models.ExternalModel) ([]byte, error) {
	var body []byte
	var err error

	if target.Format == models.FormatAnthropic {
		anthropicReq := ToAnthropic(req)
		anthropicReq.Model = target.Name
		body, err = json.Marshal(anthropicReq)
	} else {
		req.Model = target.Name
		body, err = json.Marshal(req)
	}
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, target.URL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if target.Format == models.FormatAnthropic {
		httpReq.Header.Set("x-api-key", target.APIKey)
		httpReq.Header.Set("anthropic-version", anthropicVersion)
	} else {
		httpReq.Header.Set("Authorization", "Bearer "+target.APIKey)
	}

	resp, err := f.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("forwarding request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < minSuccessStatus || resp.StatusCode >= maxSuccessStatus {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
		return nil, fmt.Errorf("external API returned %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Accumulate streaming response
	var buf bytes.Buffer
	reader := io.LimitReader(resp.Body, maxResponseSize)
	_, err = io.Copy(&buf, reader)
	if err != nil {
		return nil, fmt.Errorf("reading streaming response: %w", err)
	}

	return buf.Bytes(), nil
}

func (f *Forwarder) ForwardAnthropic(ctx context.Context, req *models.AnthropicRequest, target models.ExternalModel) ([]byte, error) {
	req.Model = target.Name
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, target.URL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", target.APIKey)
	httpReq.Header.Set("anthropic-version", anthropicVersion)

	resp, err := f.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("forwarding request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode < minSuccessStatus || resp.StatusCode >= maxSuccessStatus {
		return nil, fmt.Errorf("external API returned %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
