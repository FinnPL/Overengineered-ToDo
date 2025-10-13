package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFromEnvDefaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://root@localhost:26257/todoapp?sslmode=disable")
	cfg, err := FromEnv("userservice")
	require.NoError(t, err)
	require.Equal(t, "userservice", cfg.ServiceName)
	require.Equal(t, "8080", cfg.Port)
	require.Equal(t, 10*time.Second, cfg.ShutdownTimeout)
	require.Equal(t, "postgres://root@localhost:26257/todoapp?sslmode=disable", cfg.DatabaseURL)
}

func TestFromEnvOverrides(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://root@localhost:26257/todoapp?sslmode=verify-full")
	t.Setenv("PORT", "9090")
	t.Setenv("SHUTDOWN_TIMEOUT_SECONDS", "30")

	cfg, err := FromEnv("todoservice")
	require.NoError(t, err)
	require.Equal(t, "todoservice", cfg.ServiceName)
	require.Equal(t, "9090", cfg.Port)
	require.Equal(t, 30*time.Second, cfg.ShutdownTimeout)
	require.Equal(t, "postgres://root@localhost:26257/todoapp?sslmode=verify-full", cfg.DatabaseURL)
}

func TestFromEnvMissingDatabaseURL(t *testing.T) {
	_, err := FromEnv("userservice")
	require.Error(t, err)
}
