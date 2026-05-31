package middleware

import (
	"context"
	"net/http"
	"strings"
)

type ipKey string

const RealIPKey ipKey = "realIP"

func RealIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getRealIP(r)
		ctx := context.WithValue(r.Context(), RealIPKey, ip)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetRealIP(r *http.Request) string {
	if r == nil || r.Context() == nil {
		return ""
	}
	ip, ok := r.Context().Value(RealIPKey).(string)
	if !ok {
		return ""
	}
	return ip
}

func getRealIP(r *http.Request) string {
	// Check X-Real-IP header first
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}
	return ip
}
