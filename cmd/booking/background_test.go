package main

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func newTestApp() (*application, *bytes.Buffer) {
	var buf bytes.Buffer
	app := &application{
		logger: slog.New(slog.NewJSONHandler(&buf, nil)),
	}
	return app, &buf
}

func TestBackgroundRecoversPanics(t *testing.T) {
	t.Parallel()

	app, logs := newTestApp()

	app.background(func() {
		panic("boom")
	})

	if err := app.waitBackground(context.Background()); err != nil {
		t.Fatalf("waitBackground: %v", err)
	}

	if !strings.Contains(logs.String(), "background has panicked") {
		t.Errorf("expected panic to be logged, got logs: %s", logs.String())
	}
}

func TestWaitBackgroundWaitsForTasks(t *testing.T) {
	t.Parallel()

	app, _ := newTestApp()

	var finished atomic.Bool
	app.background(func() {
		time.Sleep(20 * time.Millisecond)
		finished.Store(true)
	})

	if err := app.waitBackground(context.Background()); err != nil {
		t.Fatalf("waitBackground: %v", err)
	}

	if !finished.Load() {
		t.Error("waitBackground returned before the task finished")
	}
}

func TestWaitBackgroundRespectsDeadline(t *testing.T) {
	t.Parallel()

	app, _ := newTestApp()

	release := make(chan struct{})
	app.background(func() {
		<-release
	})

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := app.waitBackground(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("got error %v, want context.DeadlineExceeded", err)
	}

	// Let the task finish so the test doesn't leak a goroutine.
	close(release)
	if err := app.waitBackground(context.Background()); err != nil {
		t.Fatalf("waitBackground: %v", err)
	}
}
