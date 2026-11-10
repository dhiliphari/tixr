package main

import (
	"time"

	bookingv1 "github.com/dhiliphari/tixr/gen/booking/v1"
	catalogv1 "github.com/dhiliphari/tixr/gen/catalog/v1"
)

type event struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Venue     string    `json:"venue"`
	StartsAt  time.Time `json:"starts_at"`
	CreatedAt time.Time `json:"created_at"`
}

func eventFromProto(e *catalogv1.Event) event {
	return event{
		ID:        e.GetId(),
		Name:      e.GetName(),
		Venue:     e.GetVenue(),
		StartsAt:  e.GetStartsAt().AsTime(),
		CreatedAt: e.GetCreatedAt().AsTime(),
	}
}

type seat struct {
	ID         int64  `json:"id"`
	EventID    int64  `json:"event_id"`
	Label      string `json:"label"`
	PriceCents int64  `json:"price_cents"`
	Status     string `json:"status"`
}

func seatFromProto(s *bookingv1.Seat) seat {
	status := "unknown"
	switch s.GetStatus() {
	case bookingv1.SeatStatus_SEAT_STATUS_AVAILABLE:
		status = "available"
	case bookingv1.SeatStatus_SEAT_STATUS_BOOKED:
		status = "booked"
	}

	return seat{
		ID:         s.GetId(),
		EventID:    s.GetEventId(),
		Label:      s.GetLabel(),
		PriceCents: s.GetPriceCents(),
		Status:     status,
	}
}

type bookingResponse struct {
	ID            int64     `json:"id"`
	EventID       int64     `json:"event_id"`
	SeatID        int64     `json:"seat_id"`
	CustomerEmail string    `json:"customer_email"`
	AmountCents   int64     `json:"amount_cents"`
	CreatedAt     time.Time `json:"created_at"`
}

func bookingFromProto(b *bookingv1.Booking) bookingResponse {
	return bookingResponse{
		ID:            b.GetId(),
		EventID:       b.GetEventId(),
		SeatID:        b.GetSeatId(),
		CustomerEmail: b.GetCustomerEmail(),
		AmountCents:   b.GetAmountCents(),
		CreatedAt:     b.GetCreatedAt().AsTime(),
	}
}
