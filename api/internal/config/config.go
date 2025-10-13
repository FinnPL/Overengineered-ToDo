package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds configuration shared by the HTTP services.
type Config struct {
	ServiceName     string
	Port            string
	DatabaseURL     string
	ShutdownTimeout time.Duration
}

const (
	defaultPort            = "8080"
	defaultShutdownSeconds = 10
)

// FromEnv loads service configuration using conventional environment variables.
// Recognised variables:
//   - PORT: TCP port for the HTTP listener (defaults to 8080)
//   - DATABASE_URL: PostgreSQL-compatible connection string (required)
//   - SHUTDOWN_TIMEOUT_SECONDS: graceful shutdown timeout (defaults to 10 seconds)
func FromEnv(serviceName string) (Config, error) {
	port := valueOrDefault("PORT", defaultPort)
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	timeoutSeconds := parseIntWithDefault("SHUTDOWN_TIMEOUT_SECONDS", defaultShutdownSeconds)

	return Config{
		ServiceName:     serviceName,
		Port:            port,
		DatabaseURL:     connString,
		ShutdownTimeout: time.Duration(timeoutSeconds) * time.Second,
	}, nil
}

func valueOrDefault(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func parseIntWithDefault(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			return parsed
		}
	}
	return fallback
}
