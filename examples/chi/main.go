package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/manuelarte/embeddedswagger"
	"github.com/manuelarte/embeddedswagger/examples"
)

func main() {
	logger := slog.Default()
	if err := run(logger); err != nil {
		logger.Error("Application error", slog.Any("err", err))
	}
}

func run(logger *slog.Logger) error {
	ctx := context.Background()

	r := chi.NewRouter()
	if err := embeddedswagger.Add(embeddedswagger.Config{
		OpenAPI: examples.OpenAPI,
	}, r); err != nil {
		return fmt.Errorf("failed to add embeddedswagger: %w", err)
	}

	addr := ":3000"
	logger.InfoContext(ctx, "server listening", slog.String("addr", addr))
	err := http.ListenAndServe(addr, r)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}
