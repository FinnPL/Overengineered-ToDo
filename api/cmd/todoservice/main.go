package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"

	"overengineeredtodo/internal/config"
	"overengineeredtodo/internal/database"
	"overengineeredtodo/internal/todo"
	"overengineeredtodo/pkg/httpserver"
)

const serviceName = "todo-service"

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.FromEnv(serviceName)
	if err != nil {
		logger.Error("failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	if ginMode := os.Getenv("GIN_MODE"); ginMode == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())

	repo := todo.NewRepository(pool)
	v1 := engine.Group("/v1")
	todo.RegisterRoutes(v1.Group("/todos"), repo)

	engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": serviceName})
	})

	logger.Info("starting http server", slog.String("service", serviceName), slog.String("port", cfg.Port))
	if err := httpserver.Run(ctx, engine, cfg.Port, cfg.ShutdownTimeout); err != nil {
		logger.Error("server exited with error", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("shutdown complete", slog.String("service", serviceName))
}
