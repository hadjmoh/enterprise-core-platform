package middleware

import (
	"encoding/json"
	"net/http"
	"strings"
)

// HECAuthMiddleware validates the "Splunk <token>" header
func HECAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondError(w, "No Authorization header set", 401)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Splunk" {
			respondError(w, "Invalid Authorization header format. Expected 'Splunk <token>'", 401)
			return
		}

		// TODO: specific token validation logic (check against DB/Config)
		// For MVP, we accept any token starting with "hec-"
		token := parts[1]
		if !strings.HasPrefix(token, "hec-") {
			respondError(w, "Invalid HEC Token", 403)
			return
		}

		next(w, r)
	}
}

func respondError(w http.ResponseWriter, msg string, code int) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"text": msg,
		"code": code,
	})
}
