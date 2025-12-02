package at

import (
	"encoding/json"
	"net/http"
	"strings"
)

// WriteJSON menulis response JSON
func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// GetLoginFromHeader mengambil token login dari header
func GetLoginFromHeader(r *http.Request) string {
	return r.Header.Get("Login")
}

// GetSecretFromHeader mengambil secret dari header
func GetSecretFromHeader(r *http.Request) string {
	return r.Header.Get("Secret")
}

// GetAuthorizationFromHeader mengambil authorization token dari header
func GetAuthorizationFromHeader(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return auth
}

// URLParam mengecek apakah path cocok dengan pattern dan mengekstrak parameter
func URLParam(path, pattern string) bool {
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")

	if len(pathParts) != len(patternParts) {
		return false
	}

	for i, part := range patternParts {
		if strings.HasPrefix(part, ":") {
			continue // Parameter wildcard
		}
		if part != pathParts[i] {
			return false
		}
	}
	return true
}

// GetURLParam mengambil parameter dari URL berdasarkan pattern
func GetURLParam(path, pattern, paramName string) string {
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")

	for i, part := range patternParts {
		if part == ":"+paramName && i < len(pathParts) {
			return pathParts[i]
		}
	}
	return ""
}

// GetClientIP mengambil IP address dari request
func GetClientIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}
	return strings.Split(r.RemoteAddr, ":")[0]
}