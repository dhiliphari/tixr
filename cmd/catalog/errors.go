package main

import (
	"context"
	"errors"

	"github.com/dhiliphari/tixr/internal/catalog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (app *application) toStatus(err error) error {
	switch {
	case errors.Is(err, catalog.ErrNotFound):
		return status.Error(codes.NotFound, "event not found")
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, "request cancelled")
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, "deadline exceeded")
	default:
		app.logger.Error("internal error", "error", err)
		return status.Error(codes.Internal, "internal error")
	}
}
