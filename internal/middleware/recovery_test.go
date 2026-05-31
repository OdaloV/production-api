package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecoveryCatchesPanic(t *testing.T) {
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something went wrong")
	})

	handler := Recovery(panicHandler)

	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()

	// should not crash
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Errorf("want 500, got %d", res.Code)
	}

	expected := "Internal Server Error\n"
	if res.Body.String() != expected {
		t.Errorf("want %q, got %q", expected, res.Body.String())
	}
}

func TestRecoveryPassesNormalRequest(t *testing.T) {
	normalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	handler := Recovery(normalHandler)

	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("want 200, got %d", res.Code)
	}

	if res.Body.String() != "ok" {
		t.Errorf("want 'ok', got %q", res.Body.String())
	}
}

func TestRecoveryWithDifferentPanicTypes(t *testing.T) {
	tests := []struct {
		name  string
		panic interface{}
	}{
		{"string panic", "crash"},
		{"int panic", 123},
		{"nil panic", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic(tt.panic)
			})

			handler := Recovery(panicHandler)

			req := httptest.NewRequest("GET", "/", nil)
			res := httptest.NewRecorder()

			handler.ServeHTTP(res, req)

			if res.Code != http.StatusInternalServerError {
				t.Errorf("want 500, got %d", res.Code)
			}
		})
	}
}

func TestRecoveryContinuesAfterPanic(t *testing.T) {
	count := 0

	handler := Recovery(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count++
		panic("boom")
	}))

	req := httptest.NewRequest("GET", "/", nil)

	res1 := httptest.NewRecorder()
	handler.ServeHTTP(res1, req)

	res2 := httptest.NewRecorder()
	handler.ServeHTTP(res2, req)

	if res1.Code != 500 || res2.Code != 500 {
		t.Error("should handle multiple panics")
	}
}
