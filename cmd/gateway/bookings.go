package main

import (
	"fmt"
	"net/http"
	"strings"

	bookingv1 "github.com/dhiliphari/tixr/gen/booking/v1"
)

func (app *application) createBookingHandler(w http.ResponseWriter, r *http.Request) {
	eventID, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	var input struct {
		SeatID int64  `json:"seat_id"`
		Email  string `json:"email"`
	}

	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	errs := map[string]string{}
	if input.SeatID < 1 {
		errs["seat_id"] = "must be a positive integer"
	}
	if !strings.Contains(input.Email, "@") {
		errs["email"] = "must be a valid email address"
	}
	if len(errs) > 0 {
		app.failedValidationResponse(w, r, errs)
		return
	}

	resp, err := app.booking.CreateBooking(r.Context(), &bookingv1.CreateBookingRequest{
		EventId:       eventID,
		SeatId:        input.SeatID,
		CustomerEmail: input.Email,
	})
	if err != nil {
		app.grpcErrorResponse(w, r, err)
		return
	}

	booking := bookingFromProto(resp.Booking)

	headers := http.Header{"Location": []string{fmt.Sprintf("/v1/bookings/%d", booking.ID)}}
	if err := app.writeJSON(w, http.StatusCreated, envelope{"booking": booking}, headers); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showBookingHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	resp, err := app.booking.GetBooking(r.Context(), &bookingv1.GetBookingRequest{Id: id})
	if err != nil {
		app.grpcErrorResponse(w, r, err)
		return
	}

	booking := bookingFromProto(resp.Booking)

	if err := app.writeJSON(w, http.StatusOK, envelope{"booking": booking}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
