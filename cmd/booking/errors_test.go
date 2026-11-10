package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/dhiliphari/tixr/internal/booking"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestToStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want codes.Code
	}{
		{"wrapped not found", fmt.Errorf("get booking 7: %w", booking.ErrNotFound), codes.NotFound},
		{"seat unavailable", booking.ErrSeatUnavailable, codes.FailedPrecondition},
		{"sales closed", errSalesClosed, codes.FailedPrecondition},
		{"deadline", fmt.Errorf("query: %w", context.DeadlineExceeded), codes.DeadlineExceeded},
		{"unexpected", errors.New("connection reset by peer"), codes.Internal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app, _ := newTestApp()
			if got := status.Code(app.toStatus(tt.err)); got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestToStatusDoesNotLeakInternalErrors(t *testing.T) {
	t.Parallel()

	app, _ := newTestApp()
	err := app.toStatus(errors.New(`relation "bookings" violates constraint "secret_internal_name"`))

	if msg := status.Convert(err).Message(); strings.Contains(msg, "secret_internal_name") {
		t.Errorf("internal detail leaked to the client: %q", msg)
	}
}
