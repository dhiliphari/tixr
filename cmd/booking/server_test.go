package main

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	bookingv1 "github.com/dhiliphari/tixr/gen/booking/v1"
	catalogv1 "github.com/dhiliphari/tixr/gen/catalog/v1"
)

// fakeCatalog implements catalogv1.CatalogServiceClient. The embedded
// interface is nil, so calling any method we haven't overridden panics —
// which is what we want if a test touches something unexpected.
type fakeCatalog struct {
	catalogv1.CatalogServiceClient
	resp *catalogv1.GetEventResponse
	err  error
}

func (f *fakeCatalog) GetEvent(ctx context.Context, in *catalogv1.GetEventRequest, opts ...grpc.CallOption) (*catalogv1.GetEventResponse, error) {
	return f.resp, f.err
}

func TestCreateBookingBeforeTheDatabase(t *testing.T) {
	t.Parallel()

	now := time.Date(2027, 3, 1, 12, 0, 0, 0, time.UTC)
	startedEvent := &catalogv1.GetEventResponse{
		Event: &catalogv1.Event{Id: 1, StartsAt: timestamppb.New(now.Add(-time.Hour))},
	}

	tests := []struct {
		name    string
		req     *bookingv1.CreateBookingRequest
		catalog *fakeCatalog
		want    codes.Code
	}{
		{
			name:    "invalid email",
			req:     &bookingv1.CreateBookingRequest{EventId: 1, SeatId: 1, CustomerEmail: "nope"},
			catalog: &fakeCatalog{},
			want:    codes.InvalidArgument,
		},
		{
			name:    "event not found",
			req:     &bookingv1.CreateBookingRequest{EventId: 1, SeatId: 1, CustomerEmail: "ada@example.com"},
			catalog: &fakeCatalog{err: status.Error(codes.NotFound, "event not found")},
			want:    codes.NotFound,
		},
		{
			name:    "catalog unavailable",
			req:     &bookingv1.CreateBookingRequest{EventId: 1, SeatId: 1, CustomerEmail: "ada@example.com"},
			catalog: &fakeCatalog{err: status.Error(codes.Unavailable, "connection refused")},
			want:    codes.Unavailable,
		},
		{
			name:    "catalog rejects our request",
			req:     &bookingv1.CreateBookingRequest{EventId: 1, SeatId: 1, CustomerEmail: "ada@example.com"},
			catalog: &fakeCatalog{err: status.Error(codes.InvalidArgument, "bad id")},
			want:    codes.Internal,
		},
		{
			name:    "event already started",
			req:     &bookingv1.CreateBookingRequest{EventId: 1, SeatId: 1, CustomerEmail: "ada@example.com"},
			catalog: &fakeCatalog{resp: startedEvent},
			want:    codes.FailedPrecondition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app, _ := newTestApp()
			app.catalog = tt.catalog
			app.now = func() time.Time { return now }

			_, err := app.CreateBooking(context.Background(), tt.req)
			if got := status.Code(err); got != tt.want {
				t.Errorf("got %s, want %s (error: %v)", got, tt.want, err)
			}
		})
	}
}
