// Package v1shared holds cross-endpoint HTTP concerns for API version 1: the
// error envelope and JSON response helpers.
package v1shared

import (
	"encoding/json"
	"net/http"

	"modelrail/internal/constant"
)

// WriteJSON sets the JSON content type, writes the status and encodes payload.
func WriteJSON(w http.ResponseWriter, status int, payload any) (err error) {
	w.Header().Set(constant.HeaderContentType, constant.ContentTypeJSON)
	w.WriteHeader(status)
	err = json.NewEncoder(w).Encode(payload)
	return
}
