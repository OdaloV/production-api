package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port              string
	RateLimit         int
	RateWindowSeconds int
	TimeoutSeconds    int
	AllowedOrigins    []string
}

func Load() *Config {
	return &Config{
		Port:              getEnv("PORT", "8080"),
		RateLimit:         getEnvInt("RATE_LIMIT", 100),
		RateWindowSeconds: getEnvInt("RATE_WINDOW_SECONDS", 60),
		TimeoutSeconds:    getEnvInt("TIMEOUT_SECONDS", 30),
		AllowedOrigins:    getEnvSlice("ALLOWED_ORIGINS", []string{}),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		parts := strings.Split(value, ",")
		result := make([]string, 0, len(parts))
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return defaultValue
}
