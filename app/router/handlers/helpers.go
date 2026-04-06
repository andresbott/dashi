package handlers

import (
	"encoding/json"
	"net/http"
)

// ErrorJSON writes a JSON error response with the correct Content-Type header.
func ErrorJSON(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
