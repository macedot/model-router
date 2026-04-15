package services

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"model-router/models"
)

const (
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

// buildRequest creates an HTTP request with the correct auth headers for the target format.
func (f *Forwarder) buildRequest(ctx context.Context, target models.ExternalModel, body []byte) (*http.Request, error) {
	log.Printf("[forwarder] %s → %s", target.Name, target.URL)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, target.URL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if target.Format == models.FormatOpenAI {
		httpReq.Header.Set("Authorization", "Bearer "+target.APIKey)
	} else {
		httpReq.Header.Set("x-api-key", target.APIKey)
		httpReq.Header.Set("anthropic-version", anthropicVersion)
	}

	return httpReq, nil
}

// send executes a request and returns the response body.
func (f *Forwarder) send(httpReq *http.Request) ([]byte, error) {
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

// Forward sends a prepared request body to the target provider and returns the response.
func (f *Forwarder) Forward(ctx context.Context, body []byte, target models.ExternalModel) ([]byte, error) {
	httpReq, err := f.buildRequest(ctx, target, body)
	if err != nil {
		return nil, err
	}

	return f.send(httpReq)
}

// ForwardStream sends a prepared request body to the target provider,
// writing response chunks directly to w as they arrive.
func (f *Forwarder) ForwardStream(ctx context.Context, body []byte, target models.ExternalModel, w http.ResponseWriter) error {
	httpReq, err := f.buildRequest(ctx, target, body)
	if err != nil {
		return err
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

	w.WriteHeader(resp.StatusCode)
	fw := newFlushWriter(w)
	limited := &limitedReader{R: resp.Body, Left: maxResponseSize}
	_, err = io.Copy(fw, limited)
	if err == errResponseTruncated {
		return fmt.Errorf("response exceeds maximum size limit of %d bytes", maxResponseSize)
	}
	if err != nil {
		return fmt.Errorf("reading streaming response: %w", err)
	}
	fw.Flush()

	return nil
}

// flushWriter wraps an http.ResponseWriter and flushes after every write,
// enabling true streaming for SSE/chunked responses (goproxy pattern).
type flushWriter struct {
	w  http.ResponseWriter
	fw *bufio.Writer
}

func newFlushWriter(w http.ResponseWriter) *flushWriter {
	return &flushWriter{w: w, fw: bufio.NewWriterSize(w, 4096)}
}

// Write buffers p and flushes immediately so each SSE chunk arrives at the client
// without waiting for the buffer to fill.
func (fw *flushWriter) Write(p []byte) (n int, err error) {
	n, err = fw.fw.Write(p)
	if err == nil {
		err = fw.fw.Flush()
	}
	return
}

func (fw *flushWriter) Header() http.Header {
	return fw.w.Header()
}

func (fw *flushWriter) WriteHeader(statusCode int) {
	fw.w.WriteHeader(statusCode)
}

func (fw *flushWriter) Flush() {
	fw.fw.Flush()
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
