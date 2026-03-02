package auth

import (
	"context"
	"net/http"
)

// Middleware provides common HTTP middleware.
type Middleware struct {
	sessionManager *SessionManager
}

// NewMiddleware creates a new middleware instance.
func NewMiddleware(sessionManager *SessionManager) *Middleware {
	return &Middleware{
		sessionManager: sessionManager,
	}
}

// RequireAuth middleware requires authentication.
func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionID, err := extractSessionID(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		session, err := m.sessionManager.GetSession(r.Context(), sessionID)
		if err != nil || !session.Valid() {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Add to context
		ctx := r.Context()
		ctx = contextWithUser(ctx, session.UserInfo)
		ctx = contextWithSession(ctx, session)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth middleware optionally authenticates.
func (m *Middleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionID, err := extractSessionID(r)
		if err == nil {
			session, err := m.sessionManager.GetSession(r.Context(), sessionID)
			if err == nil && session.Valid() {
				ctx := r.Context()
				ctx = contextWithUser(ctx, session.UserInfo)
				ctx = contextWithSession(ctx, session)
				r = r.WithContext(ctx)
			}
		}

		next.ServeHTTP(w, r)
	})
}

// RequireRole middleware requires specific role.
func (m *Middleware) RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := GetUserFromContext(r.Context())
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Placeholder: check user role
			_ = user
			_ = role

			next.ServeHTTP(w, r)
		})
	}
}

// CORS middleware adds CORS headers.
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// Handle preflight
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// contextWithUser adds user to context.
func contextWithUser(ctx context.Context, user *UserInfo) context.Context {
	return context.WithValue(ctx, "user", user)
}

// contextWithSession adds session to context.
func contextWithSession(ctx context.Context, session *Session) context.Context {
	return context.WithValue(ctx, "session", session)
}
