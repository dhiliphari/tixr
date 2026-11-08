package main

import (
	"context"
	"fmt"
	"runtime/debug"
)

// background runs fn in a new goroutine that the application owns. A panic in
// fn is logged instead of crashing the process, and waitBackground waits for
// fn to return.
func (app *application) background(fn func()) {
	app.wg.Add(1)

	go func() {
		defer app.wg.Done()

		defer func() {
			if pv := recover(); pv != nil {
				app.logger.Error("background has panicked",
					"panic", fmt.Sprint(pv),
					"stack", string(debug.Stack()))
			}
		}()

		fn()
	}()
}

// waitBackground blocks until all background tasks have finished, or until
// ctx is done, whichever comes first.
func (app *application) waitBackground(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		app.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("waiting for background tasks: %w", ctx.Err())
	}
}
