package orchestrate

import (
	"net/http"
	"strings"
)

// AuthMiddleware provides API key authentication.
type AuthMiddleware struct {
	apiKeys map[string]bool
}

// NewAuthMiddleware creates an auth middleware with API keys.
func NewAuthMiddleware(keys []string) *AuthMiddleware {
	m := &AuthMiddleware{
		apiKeys: make(map[string]bool),
	}
	for _, k := range keys {
		m.apiKeys[k] = true
	}
	return m
}

// Middleware wraps an HTTP handler with authentication.
func (a *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health endpoints
		if r.URL.Path == "/health" || r.URL.Path == "/ready" {
			next.ServeHTTP(w, r)
			return
		}

		// Extract API key from Authorization header
		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		// Support "Bearer <token>" or just "<token>"
		token := auth
		if strings.HasPrefix(auth, "Bearer ") {
			token = strings.TrimPrefix(auth, "Bearer ")
		}

		// Check if key is valid
		if !a.apiKeys[token] {
			http.Error(w, "invalid API key", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
