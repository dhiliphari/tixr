package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/dhiliphari/tixr/internal/booking"
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

	b, err := app.booking.Create(r.Context(), eventID, input.SeatID, input.Email)
	if err != nil {
		switch {
		case errors.Is(err, booking.ErrNotFound):
			app.notFoundResponse(w, r)
		case errors.Is(err, booking.ErrSeatUnavailable):
			app.conflictResponse(w, r, "the seat is no longer available")
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	headers := http.Header{"Location": []string{fmt.Sprintf("/v1/bookings/%d", b.ID)}}
	if err := app.writeJSON(w, http.StatusCreated, envelope{"booking": b}, headers); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showBookingHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	b, err := app.booking.Get(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, booking.ErrNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if err := app.writeJSON(w, http.StatusOK, envelope{"booking": b}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
