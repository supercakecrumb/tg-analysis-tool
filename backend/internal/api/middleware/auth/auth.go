package auth

import (
	"net/http"

	"github.com/supercakecrumb/tg-analysis-tool/internal/infrastructure/redis"
)

// RequireAuth is a middleware that checks if the user is authenticated
func RequireAuth(redis *redis.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: Implement actual authentication check
			// For now, just pass through

			// In a real implementation:
			// 1. Check for session cookie
			// 2. Validate session in Redis
			// 3. Load user context
			// 4. Pass to next handler or return 401

			next.ServeHTTP(w, r)
		})
	}
}
