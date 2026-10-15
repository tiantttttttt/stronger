package config

import (
	"fmt"
	"os"
)

// Config contains the minimum runtime configuration for the API.
type Config struct {
	Debug       bool
	HTTPAddr    string
	DatabaseURL string
	JWTSecret   string
}

// Load reads configuration from environment variables.
func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:    valueOrDefault("HTTP_ADDR", ":8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
	}
	switch valueOrDefault("DEBUG", "false") {
	case "true":
		cfg.Debug = true
	case "false":
		cfg.Debug = false
	default:
		return Config{}, fmt.Errorf("DEBUG must be true or false")
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required")
	}
	return cfg, nil
}

func valueOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
