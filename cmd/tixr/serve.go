package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error {
	srv := http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(app.logger.Handler(), slog.LevelError),
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// buffered channel to unblock goroutine even if nobody reads it
	serverErr := make(chan error, 1)
	go func() {
		app.logger.Info("starting server", "addr", srv.Addr, "env", app.config.env, "version", version)
		serverErr <- srv.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		// server failed before any signal arrived
		return fmt.Errorf("listen and serve: %w", err)
	case <-ctx.Done():
	}

	// restore the default signal behaviour
	stop()

	app.logger.Info("shutting down server", "timeout", app.config.shutdownTimeout.String())

	shutdownCtx, cancel := context.WithTimeout(context.Background(), app.config.shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	app.logger.Info("waiting for background tasks")
	if err := app.waitBackground(shutdownCtx); err != nil {
		return err
	}

	app.logger.Info("server stopped")
	return nil
}
