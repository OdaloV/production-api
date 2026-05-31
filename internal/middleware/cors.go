package middleware

import (
	"net/http"
)

// cors middleware handles cross-origin requests
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// check if origin is allowed
			if isOriginAllowed(origin, allowedOrigins) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			// set standard cors headers
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			// handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isOriginAllowed checks if origin matches any allowed pattern
func isOriginAllowed(origin string, allowed []string) bool {
	if len(allowed) == 0 {
		return false
	}

	for _, a := range allowed {
		if a == "*" {
			return true
		}
		if a == origin {
			return true
		}
	}
	return false
}
