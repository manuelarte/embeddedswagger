// Package main with an example on how to use embeddedswagger
package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/manuelarte/embeddedswagger"
)

//go:embed openapi.json
var openapi []byte

func main() {
	logger := slog.Default()
	if err := run(logger); err != nil {
		logger.Error("Application error", slog.Any("err", err))
	}
}

func run(logger *slog.Logger) error {
	ctx := context.Background()

	handler := http.NewServeMux()
	s := &http.Server{
		Addr:           ":8080",
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := embeddedswagger.Add(embeddedswagger.Config{
		OpenAPI: openapi,
	}, handler); err != nil {
		return fmt.Errorf("failed to add embeddedswagger: %w", err)
	}

	logger.InfoContext(ctx, "server listening", "addr", s.Addr)

	if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}
