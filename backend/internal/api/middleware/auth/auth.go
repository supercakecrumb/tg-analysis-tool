package auth

import (
	"net/http"

	"github.com/supercakecrumb/tg-analysis-tool/internal/infrastructure/redis"
)

// RequireAuth is a middleware that checks if the user is authenticated
func RequireAuth(redis *redis.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Session issuance is not implemented. Deny access until verified
			// server-side sessions and user scoping are available.
			http.Error(w, "Authentication is not configured", http.StatusUnauthorized)
		})
	}
}
