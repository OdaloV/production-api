package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHealthCheckReturns200(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	res := httptest.NewRecorder()

	HealthCheck(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("want status 200, got %d", res.Code)
	}
}

func TestHealthCheckReturnsJSON(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	res := httptest.NewRecorder()

	HealthCheck(res, req)

	contentType := res.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("want 'application/json', got '%s'", contentType)
	}
}

func TestHealthCheckHasStatusOk(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	res := httptest.NewRecorder()

	HealthCheck(res, req)

	var response healthResponse
	err := json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		t.Fatalf("failed to parse json: %v", err)
	}

	if response.Status != "ok" {
		t.Errorf("want status 'ok', got '%s'", response.Status)
	}
}

func TestHealthCheckHasTimestamp(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	res := httptest.NewRecorder()

	HealthCheck(res, req)

	var response healthResponse
	err := json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		t.Fatalf("failed to parse json: %v", err)
	}

	if response.Timestamp == "" {
		t.Error("timestamp should not be empty")
	}
}

func TestHealthCheckTimestampIsValid(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	res := httptest.NewRecorder()

	HealthCheck(res, req)

	var response healthResponse
	err := json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		t.Fatalf("failed to parse json: %v", err)
	}

	_, err = time.Parse(time.RFC3339, response.Timestamp)
	if err != nil {
		t.Errorf("timestamp '%s' is not valid RFC3339 format: %v", response.Timestamp, err)
	}
}

func TestHealthCheckTimestampIsRecent(t *testing.T) {
	before := time.Now().UTC().Add(-1 * time.Second)

	req := httptest.NewRequest("GET", "/health", nil)
	res := httptest.NewRecorder()

	HealthCheck(res, req)

	var response healthResponse
	err := json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		t.Fatalf("failed to parse json: %v", err)
	}

	timestamp, err := time.Parse(time.RFC3339, response.Timestamp)
	if err != nil {
		t.Fatalf("invalid timestamp: %v", err)
	}

	if timestamp.Before(before) {
		t.Errorf("timestamp '%s' is not recent", response.Timestamp)
	}
}

func TestHealthCheckResponseStructure(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	res := httptest.NewRecorder()

	HealthCheck(res, req)

	var response map[string]interface{}
	err := json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		t.Fatalf("failed to parse json: %v", err)
	}

	if _, ok := response["status"]; !ok {
		t.Error("response missing 'status' field")
	}

	if _, ok := response["timestamp"]; !ok {
		t.Error("response missing 'timestamp' field")
	}
}
