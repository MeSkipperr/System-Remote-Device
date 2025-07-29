package api

import (
	"encoding/json"
	"net/http"
)

// WriteJSON is a helper function to send JSON responses
func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}