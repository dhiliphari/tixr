package main

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	bookingv1 "github.com/dhiliphari/tixr/gen/booking/v1"
	"github.com/dhiliphari/tixr/internal/booking"
)

func toProtoSeat(s booking.Seat) *bookingv1.Seat {
	return &bookingv1.Seat{
		Id:         s.ID,
		EventId:    s.EventID,
		Label:      s.Label,
		PriceCents: int64(s.PriceCents),
		Status:     toProtoSeatStatus(s.Status),
	}
}

func toProtoSeatStatus(s string) bookingv1.SeatStatus {
	switch s {
	case booking.SeatAvailable:
		return bookingv1.SeatStatus_SEAT_STATUS_AVAILABLE
	case booking.SeatBooked:
		return bookingv1.SeatStatus_SEAT_STATUS_BOOKED
	default:
		return bookingv1.SeatStatus_SEAT_STATUS_UNSPECIFIED
	}
}

func toProtoBooking(b booking.Booking) *bookingv1.Booking {
	return &bookingv1.Booking{
		Id:            b.ID,
		EventId:       b.EventID,
		SeatId:        b.SeatID,
		CustomerEmail: b.CustomerEmail,
		AmountCents:   int64(b.AmountCents),
		CreatedAt:     timestamppb.New(b.CreatedAt),
	}
}
