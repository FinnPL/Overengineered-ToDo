package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Run starts the http.Server and blocks until the context is cancelled or the server stops.
// It enables graceful shutdown using the supplied timeout.
func Run(ctx context.Context, handler http.Handler, port string, shutdownTimeout time.Duration) error {
	server := &http.Server{
		Addr:              fmt.Sprintf(":%s", port),
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown server: %w", err)
		}
		return <-errCh
	case err := <-errCh:
		return err
	}
}
