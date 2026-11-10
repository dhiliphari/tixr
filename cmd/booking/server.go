package main

import (
	"context"
	"errors"
	"strings"
	"time"

	bookingv1 "github.com/dhiliphari/tixr/gen/booking/v1"
	catalogv1 "github.com/dhiliphari/tixr/gen/catalog/v1"
	"github.com/dhiliphari/tixr/internal/booking"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var errSalesClosed = errors.New("sales for this event have closed")

func (app *application) ListSeats(ctx context.Context, req *bookingv1.ListSeatsRequest) (*bookingv1.ListSeatsResponse, error) {
	seats, err := app.store.ListSeats(ctx, req.GetEventId())
	if err != nil {
		return nil, app.toStatus(err)
	}

	resp := &bookingv1.ListSeatsResponse{
		Seats: make([]*bookingv1.Seat, 0, len(seats)),
	}
	for _, s := range seats {
		resp.Seats = append(resp.Seats, toProtoSeat(s))
	}

	return resp, nil
}

func (app *application) CreateBooking(ctx context.Context, req *bookingv1.CreateBookingRequest) (*bookingv1.CreateBookingResponse, error) {
	if req.GetEventId() < 1 || req.GetSeatId() < 1 || !strings.Contains(req.GetCustomerEmail(), "@") {
		return nil, status.Error(codes.InvalidArgument, "event_id, seat_id and a valid customer_email are required")
	}

	ev, err := app.catalog.GetEvent(ctx, &catalogv1.GetEventRequest{Id: req.GetEventId()})
	if err != nil {
		return nil, app.fromCatalog(err, req.GetEventId())
	}

	if !ev.GetEvent().GetStartsAt().AsTime().After(app.now()) {
		return nil, app.toStatus(errSalesClosed)
	}

	b, err := app.store.Create(ctx, req.GetEventId(), req.GetSeatId(), req.GetCustomerEmail())
	if err != nil {
		return nil, app.toStatus(err)
	}

	app.sendConfirmation(ctx, b)

	return &bookingv1.CreateBookingResponse{Booking: toProtoBooking(b)}, nil
}

func (app *application) GetBooking(ctx context.Context, req *bookingv1.GetBookingRequest) (*bookingv1.GetBookingResponse, error) {
	b, err := app.store.Get(ctx, req.GetId())
	if err != nil {
		return nil, app.toStatus(err)
	}

	return &bookingv1.GetBookingResponse{Booking: toProtoBooking(b)}, nil
}

// sendConfirmation emails the customer in the background.
// The RPC's context is cancelled when the handler returns, so the task keeps
// its values but not its cancellation.
func (app *application) sendConfirmation(ctx context.Context, b booking.Booking) {
	reqCtx := context.WithoutCancel(ctx)

	app.background(func() {
		ctx, cancel := context.WithTimeout(reqCtx, 10*time.Second)
		defer cancel()

		if err := app.mailer.SendBookingConfirmation(ctx, b); err != nil {
			app.logger.Error("send booking confirmation", "error", err, "booking_id", b.ID)
		}
	})
}
