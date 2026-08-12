package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ViniciusToledoNunes/sezzle-calculator/backend/internal/httpapi"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	server := newServer(logger)

	logger.Info("calculator API listening", "address", server.Addr)
	if err := run(context.Background(), logger, server); err != nil {
		logger.Error("server stopped unexpectedly", "error", err)
		os.Exit(1)
	}
}

func newServer(logger *slog.Logger) *http.Server {
	return &http.Server{
		Addr:              ":" + envOrDefault("PORT", "8080"),
		Handler:           httpapi.NewHandler(logger, envOrDefault("ALLOWED_ORIGIN", "http://localhost:5173")),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

func run(parent context.Context, logger *slog.Logger, server *http.Server) error {
	shutdownSignal, stop := signal.NotifyContext(parent, syscall.SIGINT, syscall.SIGTERM)
	shutdownComplete := make(chan struct{})

	go func() {
		defer close(shutdownComplete)
		<-shutdownSignal.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			logger.Error("graceful shutdown failed", "error", err)
		}
	}()

	err := server.ListenAndServe()
	stop()
	<-shutdownComplete
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
