package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"model-router/config"
)

// decodeWithLimit decodes a JSON request body with a size limit.
// It uses config.Defaults.BodyLimit (50MB) to prevent memory exhaustion
// from large request bodies.
func decodeWithLimit(r *http.Request, v interface{}) error {
	limited := io.LimitReader(r.Body, int64(config.Defaults.BodyLimit))
	if err := json.NewDecoder(limited).Decode(v); err != nil {
		// Check if the error is from LimitReader
		// io.LimitReader returns EOF when the limit is exceeded
		return err
	}
	return nil
}
