package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecoverMiddleware(t *testing.T) {
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something broke")
	})

	handler := &recoverMiddleware{handler: panicHandler}

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusInternalServerError)
	}

	var errResp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &errResp)
	errMap := errResp["error"].(map[string]interface{})
	if errMap["message"] != "Internal server error" {
		t.Errorf("message = %q; want %q", errMap["message"], "Internal server error")
	}
}

func TestRecoverMiddleware_NoPanic(t *testing.T) {
	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	})

	handler := &recoverMiddleware{handler: okHandler}

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != `{"ok":true}` {
		t.Errorf("body = %q; want %q", rec.Body.String(), `{"ok":true}`)
	}
}
