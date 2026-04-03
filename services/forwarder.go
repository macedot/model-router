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
		openAIReq := *req
		openAIReq.Model = target.Name
		body, err = json.Marshal(&openAIReq)
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

// ForwardOpenAIStream forwards a streaming OpenAI request to the target provider,
// writing chunks directly to w as they arrive. It enforces maxResponseSize via limitedReader.
func (f *Forwarder) ForwardOpenAIStream(ctx context.Context, req *models.OpenAIRequest, target models.ExternalModel, w io.Writer) error {
	var body []byte
	var err error

	if target.Format == models.FormatAnthropic {
		anthropicReq := ToAnthropic(req)
		anthropicReq.Model = target.Name
		body, err = json.Marshal(anthropicReq)
	} else {
		openAIReq := *req
		openAIReq.Model = target.Name
		body, err = json.Marshal(&openAIReq)
	}
	if err != nil {
		return fmt.Errorf("marshaling request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, target.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
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
		return fmt.Errorf("forwarding request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < minSuccessStatus || resp.StatusCode >= maxSuccessStatus {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
		return fmt.Errorf("external API returned %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Stream response directly to writer with size limit.
	limited := &limitedReader{
		R:    resp.Body,
		Left: maxResponseSize,
	}
	_, err = io.Copy(w, limited)
	if err == errResponseTruncated {
		return fmt.Errorf("response exceeds maximum size limit of %d bytes", maxResponseSize)
	}
	if err != nil {
		return fmt.Errorf("reading streaming response: %w", err)
	}

	return nil
}

// errResponseTruncated is returned when a response exceeds the size limit.
var errResponseTruncated = fmt.Errorf("response truncated")

// limitedReader wraps a reader and tracks remaining bytes.
// It returns errResponseTruncated when the limit is exceeded.
type limitedReader struct {
	R    io.Reader
	Left int64
}

func (l *limitedReader) Read(p []byte) (int, error) {
	if l.Left <= 0 {
		return 0, errResponseTruncated
	}
	if int64(len(p)) > l.Left {
		p = p[:l.Left]
	}
	n, err := l.R.Read(p)
	l.Left -= int64(n)
	if err == io.EOF && l.Left <= 0 {
		return n, nil
	}
	if l.Left <= 0 && err == nil {
		return n, errResponseTruncated
	}
	return n, err
}

func (f *Forwarder) ForwardAnthropic(ctx context.Context, req *models.AnthropicRequest, target models.ExternalModel) ([]byte, error) {
	anthropicReq := *req
	anthropicReq.Model = target.Name
	body, err := json.Marshal(&anthropicReq)
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
