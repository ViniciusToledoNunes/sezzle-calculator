package main

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"testing"
	"time"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestEnvOrDefault(t *testing.T) {
	t.Setenv("CALCULATOR_TEST_VALUE", "configured")
	if got := envOrDefault("CALCULATOR_TEST_VALUE", "fallback"); got != "configured" {
		t.Errorf("envOrDefault() = %q, want configured", got)
	}

	if got := envOrDefault("CALCULATOR_MISSING_VALUE", "fallback"); got != "fallback" {
		t.Errorf("envOrDefault() = %q, want fallback", got)
	}
}

func TestNewServerConfiguration(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("ALLOWED_ORIGIN", "https://calculator.example")

	server := newServer(testLogger())
	if server.Addr != ":9090" {
		t.Errorf("Addr = %q, want :9090", server.Addr)
	}
	if server.Handler == nil {
		t.Error("Handler must be configured")
	}
	if server.ReadHeaderTimeout != 5*time.Second || server.ReadTimeout != 10*time.Second || server.WriteTimeout != 10*time.Second || server.IdleTimeout != 60*time.Second {
		t.Error("server timeouts do not match the production configuration")
	}
}

func TestRunStopsGracefully(t *testing.T) {
	started := make(chan struct{})
	server := &http.Server{
		Addr:    "127.0.0.1:0",
		Handler: http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}),
		BaseContext: func(net.Listener) context.Context {
			close(started)
			return context.Background()
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() { result <- run(ctx, testLogger(), server) }()

	select {
	case <-started:
		cancel()
	case <-time.After(2 * time.Second):
		t.Fatal("server did not start")
	}

	select {
	case err := <-result:
		if err != nil {
			t.Fatalf("run() error = %v, want nil", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server did not stop gracefully")
	}
}

func TestRunReturnsListenError(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	server := &http.Server{Addr: listener.Addr().String()}
	if err := run(context.Background(), testLogger(), server); err == nil {
		t.Fatal("run() error = nil, want address-in-use error")
	}
}
