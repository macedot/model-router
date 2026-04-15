package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteError_WithErr(t *testing.T) {
	rec := httptest.NewRecorder()

	writeError(rec, http.StatusBadRequest, "Invalid request", "invalid_request_error", errors.New("json decode failed"))

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)

	errObj := resp["error"].(map[string]interface{})
	if errObj["message"] != "Invalid request" {
		t.Errorf("message = %q, want %q", errObj["message"], "Invalid request")
	}
	if errObj["type"] != "invalid_request_error" {
		t.Errorf("type = %q, want %q", errObj["type"], "invalid_request_error")
	}
}

func TestWriteError_NilErr(t *testing.T) {
	rec := httptest.NewRecorder()

	writeError(rec, http.StatusBadGateway, "All providers failed", "api_error", nil)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadGateway)
	}

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)

	errObj := resp["error"].(map[string]interface{})
	if errObj["message"] != "All providers failed" {
		t.Errorf("message = %q", errObj["message"])
	}
}

func TestWriteError_ContentType(t *testing.T) {
	rec := httptest.NewRecorder()

	writeError(rec, http.StatusInternalServerError, "Internal error", "api_error", nil)

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}
}

func TestWriteError_SanitizedMessage(t *testing.T) {
	rec := httptest.NewRecorder()

	// Internal error should NOT leak to client
	writeError(rec, http.StatusInternalServerError, "Internal server error", "api_error", errors.New("database connection string: postgres://admin:secret@db"))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	body := rec.Body.String()
	// The client message should be present
	if !contains(body, "Internal server error") {
		t.Errorf("response missing sanitized message, got: %s", body)
	}
	// The internal error should NOT leak
	if contains(body, "postgres://admin:secret@db") {
		t.Error("internal error details leaked to client")
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
