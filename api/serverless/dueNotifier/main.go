package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jackc/pgx/v5/pgxpool"

	"overengineeredtodo/internal/database"
	"overengineeredtodo/internal/todo"
)

var (
	pool     *pgxpool.Pool
	poolOnce sync.Once
	poolErr  error
	logger   = slog.New(slog.NewJSONHandler(os.Stdout, nil))
)

// Event represents the input payload for the Lambda invocation.
type Event struct {
	WindowMinutes int `json:"window_minutes"`
}

// Response provides a lightweight summary of upcoming todos.
type Response struct {
	WindowMinutes int         `json:"window_minutes"`
	Count         int         `json:"count"`
	Todos         []todo.Todo `json:"todos"`
}

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, event Event) (Response, error) {
	if event.WindowMinutes <= 0 {
		event.WindowMinutes = 60
	}

	pool, err := getPool(ctx)
	if err != nil {
		logger.Error("failed to initialise database pool", slog.String("error", err.Error()))
		return Response{}, err
	}

	repo := todo.NewRepository(pool)
	window := time.Duration(event.WindowMinutes) * time.Minute
	todos, err := repo.ListDueWithin(ctx, window)
	if err != nil {
		logger.Error("list due todos failed", slog.String("error", err.Error()))
		return Response{}, err
	}

	logger.Info("found due todos", slog.Int("count", len(todos)), slog.Int("window_minutes", event.WindowMinutes))

	return Response{
		WindowMinutes: event.WindowMinutes,
		Count:         len(todos),
		Todos:         todos,
	}, nil
}

func getPool(ctx context.Context) (*pgxpool.Pool, error) {
	poolOnce.Do(func() {
		conn := os.Getenv("DATABASE_URL")
		if conn == "" {
			poolErr = errors.New("DATABASE_URL is required")
			return
		}
		pool, poolErr = database.NewPool(ctx, conn)
	})
	return pool, poolErr
}
