package server

import (
	"encoding/json"
	"net/http"
	"strings"
)

// WriteJSON ...
func WriteJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// ReadJSON ...
func ReadJSON[T any](w http.ResponseWriter, r *http.Request, dst *T) bool {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return false
	}
	return true
}

// GetUserID ...
func GetUserID(r *http.Request) string {
	// Temporary: read from header
	uid := r.Header.Get("X-User-ID")
	return strings.TrimSpace(uid)
}
