package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCORSAddsHeadersForAllowedOrigin(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	allowedOrigins := []string{"https://example.com", "https://app.com"}
	cors := CORS(allowedOrigins)
	wrapped := cors(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://example.com")
	res := httptest.NewRecorder()

	wrapped.ServeHTTP(res, req)

	if res.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Errorf("want 'https://example.com', got '%s'", res.Header().Get("Access-Control-Allow-Origin"))
	}

	if res.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("expected Access-Control-Allow-Methods header")
	}

	if res.Header().Get("Access-Control-Allow-Headers") == "" {
		t.Error("expected Access-Control-Allow-Headers header")
	}
}

func TestCORSAddsHeadersForWildcard(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cors := CORS([]string{"*"})
	wrapped := cors(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://any-origin.com")
	res := httptest.NewRecorder()

	wrapped.ServeHTTP(res, req)

	if res.Header().Get("Access-Control-Allow-Origin") != "https://any-origin.com" {
		t.Errorf("want 'https://any-origin.com', got '%s'", res.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORSNoHeaderForDisallowedOrigin(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cors := CORS([]string{"https://allowed.com"})
	wrapped := cors(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://evil.com")
	res := httptest.NewRecorder()

	wrapped.ServeHTTP(res, req)

	if res.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("expected no header, got '%s'", res.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORSPreflightReturns204(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for preflight")
	})

	cors := CORS([]string{"https://example.com"})
	wrapped := cors(handler)

	req := httptest.NewRequest("OPTIONS", "/", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	res := httptest.NewRecorder()

	wrapped.ServeHTTP(res, req)

	if res.Code != http.StatusNoContent {
		t.Errorf("want 204, got %d", res.Code)
	}

	if res.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Errorf("want 'https://example.com', got '%s'", res.Header().Get("Access-Control-Allow-Origin"))
	}

	if res.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("expected Access-Control-Allow-Methods header")
	}
}

func TestCORSAllowsMethods(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cors := CORS([]string{"*"})
	wrapped := cors(handler)

	methods := []string{"GET", "POST", "PUT", "DELETE"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/", nil)
			req.Header.Set("Origin", "https://test.com")
			res := httptest.NewRecorder()

			wrapped.ServeHTTP(res, req)

			allowed := res.Header().Get("Access-Control-Allow-Methods")
			if !strings.Contains(allowed, method) {
				t.Errorf("allowed methods '%s' should contain '%s'", allowed, method)
			}

			if res.Code != http.StatusOK {
				t.Errorf("want 200, got %d", res.Code)
			}
		})
	}
}

func TestCORSAllowsHeaders(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cors := CORS([]string{"*"})
	wrapped := cors(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://test.com")
	res := httptest.NewRecorder()

	wrapped.ServeHTTP(res, req)

	allowedHeaders := res.Header().Get("Access-Control-Allow-Headers")
	expected := []string{"Content-Type", "Authorization"}

	for _, h := range expected {
		if !strings.Contains(allowedHeaders, h) {
			t.Errorf("expected '%s' in '%s'", h, allowedHeaders)
		}
	}
}

func TestCORSMultipleOrigins(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	allowed := []string{"https://a.com", "https://b.com", "https://c.com"}
	cors := CORS(allowed)
	wrapped := cors(handler)

	for _, origin := range allowed {
		t.Run(origin, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("Origin", origin)
			res := httptest.NewRecorder()

			wrapped.ServeHTTP(res, req)

			if res.Header().Get("Access-Control-Allow-Origin") != origin {
				t.Errorf("want '%s', got '%s'", origin, res.Header().Get("Access-Control-Allow-Origin"))
			}
		})
	}
}

func TestCORSEmptyAllowedOrigins(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cors := CORS([]string{})
	wrapped := cors(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://example.com")
	res := httptest.NewRecorder()

	wrapped.ServeHTTP(res, req)

	if res.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("expected no allow-origin header for empty allowed list")
	}
}
