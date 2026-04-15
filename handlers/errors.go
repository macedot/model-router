package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

// writeError sanitizes error messages sent to HTTP clients.
// It logs the full error server-side but writes only clientMsg to the response.
func writeError(w http.ResponseWriter, status int, clientMsg string, errorType string, err error) {
	if err != nil {
		log.Printf("[handlers] error: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		} `json:"error"`
	}{
		Error: struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		}{
			Message: clientMsg,
			Type:    errorType,
		},
	}
	json.NewEncoder(w).Encode(resp)
}
