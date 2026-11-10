package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/dhiliphari/tixr/internal/booking"
)

type mailer struct {
	logger *slog.Logger
	delay  time.Duration
}

func (m *mailer) SendBookingConfirmation(ctx context.Context, b booking.Booking) error {
	select {
	case <-time.After(m.delay):
	case <-ctx.Done():
		return fmt.Errorf("send confirmation for booking %d: %w", b.ID, ctx.Err())
	}

	m.logger.Info("confirmation email sent", "booking_id", b.ID)

	return nil
}
