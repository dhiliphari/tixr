package main

import (
	"context"
	"errors"

	"github.com/dhiliphari/tixr/internal/booking"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (app *application) toStatus(err error) error {
	switch {
	case errors.Is(err, booking.ErrNotFound):
		return status.Error(codes.NotFound, "not found")
	case errors.Is(err, booking.ErrSeatUnavailable):
		return status.Error(codes.FailedPrecondition, "the seat is no longer available")
	case errors.Is(err, errSalesClosed):
		return status.Error(codes.FailedPrecondition, errSalesClosed.Error())
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, "request cancelled")
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, "deadline exceeded")
	default:
		app.logger.Error("internal error", "error", err)
		return status.Error(codes.Internal, "internal error")
	}
}

// fromCatalog translates an error from the catalog service into a status that
// describes this service, rather than passing catalog's status through.
func (app *application) fromCatalog(err error, eventID int64) error {
	switch status.Code(err) {
	case codes.NotFound:
		return status.Error(codes.NotFound, "event not found")
	case codes.Unavailable, codes.DeadlineExceeded:
		app.logger.Warn("catalog unavailable", "error", err, "event_id", eventID)
		return status.Error(codes.Unavailable, "the event catalog is temporarily unavailable")
	default:
		app.logger.Error("unexpected error from catalog", "error", err, "event_id", eventID)
		return status.Error(codes.Internal, "internal error")
	}
}
