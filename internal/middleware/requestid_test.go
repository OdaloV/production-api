package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestID_GeneratesNewID(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := GetRequestID(r.Context())
		if id == "" {
			t.Error("Expected request ID to be set, got empty string")
		}
		w.WriteHeader(http.StatusOK)
	})

	// Wrap with middleware
	handler := RequestID(testHandler)

	// Create a request with no existing ID
	req := httptest.NewRequest("GET", "/", nil)
	recorder := httptest.NewRecorder()

	// Serve request
	handler.ServeHTTP(recorder, req)

	// Check that response header contains request ID
	responseID := recorder.Header().Get("X-Request-ID")
	if responseID == "" {
		t.Error("Expected X-Request-ID header to be set")
	}

	// Check status code
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", recorder.Code)
	}
}

func TestRequestID_RespectsExistingHeader(t *testing.T) {
	existingID := "my-existing-id-123"

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := GetRequestID(r.Context())
		if id != existingID {
			t.Errorf("Expected request ID %s, got %s", existingID, id)
		}
		w.WriteHeader(http.StatusOK)
	})

	handler := RequestID(testHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Request-ID", existingID)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	responseID := recorder.Header().Get("X-Request-ID")
	if responseID != existingID {
		t.Errorf("Expected response header %s, got %s", existingID, responseID)
	}
}

func TestGetRequestID_ReturnsEmptyWhenNotSet(t *testing.T) {
	ctx := context.Background()
	id := GetRequestID(ctx)
	if id != "" {
		t.Errorf("Expected empty string, got %s", id)
	}
}
