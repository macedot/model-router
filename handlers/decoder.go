package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"model-router/config"
	"model-router/models"
)

// readBody reads and parses the request body.
// Returns the full body as a generic map for passthrough,
// and a RequestEnvelope for validation/routing.
func readBody(r *http.Request) (map[string]interface{}, *models.RequestEnvelope, error) {
	limited := io.LimitReader(r.Body, int64(config.Defaults.BodyLimit))
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, nil, err
	}

	var body map[string]interface{}
	if err := json.Unmarshal(data, &body); err != nil {
		return nil, nil, err
	}

	var envelope models.RequestEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, nil, err
	}

	return body, &envelope, nil
}
