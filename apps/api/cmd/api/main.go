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

	"api/internal/config"
	"api/internal/server"
	"api/internal/store"
)

func main() {
	os.Exit(run())
}

func run() int {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("invalid configuration", "error", err)
		return 1
	}

	if err := os.MkdirAll(cfg.MediaDir(), 0o750); err != nil {
		slog.Error("failed to create data dirs", "error", err)
		return 1
	}

	st, err := store.Open(cfg.DBPath())
	if err != nil {
		slog.Error("failed to open store", "error", err)
		return 1
	}
	defer func() { _ = st.Close() }()

	r := server.NewRouter(cfg, st)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       60 * time.Second,
		// WriteTimeout must cover streaming video files up to 1 GiB on slow
		// connections, not just API responses.
		WriteTimeout: 10 * time.Minute,
		IdleTimeout:  120 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serveErr := make(chan error, 1)
	go func() {
		slog.Info("listening", "addr", srv.Addr)
		serveErr <- srv.ListenAndServe()
	}()

	select {
	case err := <-serveErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
			return 1
		}
	case <-ctx.Done():
		// Shutdown in the main flow and wait for it so in-flight requests
		// finish before the process exits.
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("graceful shutdown", "error", err)
		}
		if err := <-serveErr; err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
			return 1
		}
	}
	return 0
}
