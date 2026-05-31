package config

import (
	"os"
	"testing"
)

func TestLoadWithDefaults(t *testing.T) {
	// clear env vars
	os.Unsetenv("PORT")
	os.Unsetenv("RATE_LIMIT")
	os.Unsetenv("RATE_WINDOW_SECONDS")
	os.Unsetenv("TIMEOUT_SECONDS")
	os.Unsetenv("ALLOWED_ORIGINS")

	cfg := Load()

	if cfg.Port != "8080" {
		t.Errorf("want port '8080', got '%s'", cfg.Port)
	}
	if cfg.RateLimit != 100 {
		t.Errorf("want rate limit 100, got %d", cfg.RateLimit)
	}
	if cfg.RateWindowSeconds != 60 {
		t.Errorf("want rate window 60, got %d", cfg.RateWindowSeconds)
	}
	if cfg.TimeoutSeconds != 30 {
		t.Errorf("want timeout 30, got %d", cfg.TimeoutSeconds)
	}
	if len(cfg.AllowedOrigins) != 0 {
		t.Errorf("want empty allowed origins, got %v", cfg.AllowedOrigins)
	}
}

func TestLoadWithEnvVars(t *testing.T) {
	os.Setenv("PORT", "9090")
	os.Setenv("RATE_LIMIT", "200")
	os.Setenv("RATE_WINDOW_SECONDS", "120")
	os.Setenv("TIMEOUT_SECONDS", "60")
	os.Setenv("ALLOWED_ORIGINS", "https://example.com,https://app.com")

	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("RATE_LIMIT")
		os.Unsetenv("RATE_WINDOW_SECONDS")
		os.Unsetenv("TIMEOUT_SECONDS")
		os.Unsetenv("ALLOWED_ORIGINS")
	}()

	cfg := Load()

	if cfg.Port != "9090" {
		t.Errorf("want port '9090', got '%s'", cfg.Port)
	}
	if cfg.RateLimit != 200 {
		t.Errorf("want rate limit 200, got %d", cfg.RateLimit)
	}
	if cfg.RateWindowSeconds != 120 {
		t.Errorf("want rate window 120, got %d", cfg.RateWindowSeconds)
	}
	if cfg.TimeoutSeconds != 60 {
		t.Errorf("want timeout 60, got %d", cfg.TimeoutSeconds)
	}
	if len(cfg.AllowedOrigins) != 2 {
		t.Errorf("want 2 origins, got %d", len(cfg.AllowedOrigins))
	}
	if cfg.AllowedOrigins[0] != "https://example.com" {
		t.Errorf("want 'https://example.com', got '%s'", cfg.AllowedOrigins[0])
	}
	if cfg.AllowedOrigins[1] != "https://app.com" {
		t.Errorf("want 'https://app.com', got '%s'", cfg.AllowedOrigins[1])
	}
}

func TestLoadWithPartialEnvVars(t *testing.T) {
	os.Setenv("PORT", "3000")
	os.Setenv("RATE_LIMIT", "50")

	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("RATE_LIMIT")
	}()

	cfg := Load()

	if cfg.Port != "3000" {
		t.Errorf("want port '3000', got '%s'", cfg.Port)
	}
	if cfg.RateLimit != 50 {
		t.Errorf("want rate limit 50, got %d", cfg.RateLimit)
	}
	if cfg.RateWindowSeconds != 60 {
		t.Errorf("want rate window 60 (default), got %d", cfg.RateWindowSeconds)
	}
	if cfg.TimeoutSeconds != 30 {
		t.Errorf("want timeout 30 (default), got %d", cfg.TimeoutSeconds)
	}
}

func TestLoadWithInvalidIntValues(t *testing.T) {
	os.Setenv("RATE_LIMIT", "notanumber")
	os.Setenv("RATE_WINDOW_SECONDS", "invalid")
	os.Setenv("TIMEOUT_SECONDS", "invalid")

	defer func() {
		os.Unsetenv("RATE_LIMIT")
		os.Unsetenv("RATE_WINDOW_SECONDS")
		os.Unsetenv("TIMEOUT_SECONDS")
	}()

	cfg := Load()

	// should fall back to defaults
	if cfg.RateLimit != 100 {
		t.Errorf("want rate limit 100 (default), got %d", cfg.RateLimit)
	}
	if cfg.RateWindowSeconds != 60 {
		t.Errorf("want rate window 60 (default), got %d", cfg.RateWindowSeconds)
	}
	if cfg.TimeoutSeconds != 30 {
		t.Errorf("want timeout 30 (default), got %d", cfg.TimeoutSeconds)
	}
}

func TestLoadWithSingleOrigin(t *testing.T) {
	os.Setenv("ALLOWED_ORIGINS", "https://single.com")

	defer func() {
		os.Unsetenv("ALLOWED_ORIGINS")
	}()

	cfg := Load()

	if len(cfg.AllowedOrigins) != 1 {
		t.Errorf("want 1 origin, got %d", len(cfg.AllowedOrigins))
	}
	if cfg.AllowedOrigins[0] != "https://single.com" {
		t.Errorf("want 'https://single.com', got '%s'", cfg.AllowedOrigins[0])
	}
}

func TestLoadWithEmptyOrigins(t *testing.T) {
	os.Setenv("ALLOWED_ORIGINS", "")

	defer func() {
		os.Unsetenv("ALLOWED_ORIGINS")
	}()

	cfg := Load()

	if len(cfg.AllowedOrigins) != 0 {
		t.Errorf("want 0 origins, got %d", len(cfg.AllowedOrigins))
	}
}

func TestLoadWithSpacesInOrigins(t *testing.T) {
	os.Setenv("ALLOWED_ORIGINS", " https://a.com , https://b.com ")

	defer func() {
		os.Unsetenv("ALLOWED_ORIGINS")
	}()

	cfg := Load()

	if len(cfg.AllowedOrigins) != 2 {
		t.Errorf("want 2 origins, got %d", len(cfg.AllowedOrigins))
	}
	if cfg.AllowedOrigins[0] != "https://a.com" {
		t.Errorf("want 'https://a.com', got '%s'", cfg.AllowedOrigins[0])
	}
	if cfg.AllowedOrigins[1] != "https://b.com" {
		t.Errorf("want 'https://b.com', got '%s'", cfg.AllowedOrigins[1])
	}
}

func TestGetEnv(t *testing.T) {
	os.Setenv("TEST_KEY", "testvalue")
	defer os.Unsetenv("TEST_KEY")

	if got := getEnv("TEST_KEY", "default"); got != "testvalue" {
		t.Errorf("want 'testvalue', got '%s'", got)
	}

	if got := getEnv("MISSING_KEY", "default"); got != "default" {
		t.Errorf("want 'default', got '%s'", got)
	}
}

func TestGetEnvInt(t *testing.T) {
	os.Setenv("TEST_INT", "123")
	defer os.Unsetenv("TEST_INT")

	if got := getEnvInt("TEST_INT", 0); got != 123 {
		t.Errorf("want 123, got %d", got)
	}

	if got := getEnvInt("MISSING_INT", 999); got != 999 {
		t.Errorf("want 999, got %d", got)
	}

	os.Setenv("TEST_INVALID", "notnumber")
	defer os.Unsetenv("TEST_INVALID")

	if got := getEnvInt("TEST_INVALID", 42); got != 42 {
		t.Errorf("want 42 (default), got %d", got)
	}
}

func TestGetEnvSlice(t *testing.T) {
	os.Setenv("TEST_SLICE", "a,b,c")
	defer os.Unsetenv("TEST_SLICE")

	got := getEnvSlice("TEST_SLICE", []string{})
	if len(got) != 3 {
		t.Errorf("want 3 items, got %d", len(got))
	}

	os.Setenv("TEST_EMPTY", "")
	defer os.Unsetenv("TEST_EMPTY")

	got = getEnvSlice("TEST_EMPTY", []string{"default"})
	if len(got) != 1 || got[0] != "default" {
		t.Errorf("want ['default'], got %v", got)
	}

	got = getEnvSlice("MISSING_SLICE", []string{"fallback"})
	if len(got) != 1 || got[0] != "fallback" {
		t.Errorf("want ['fallback'], got %v", got)
	}
}
