package main

import (
	"errors"
	"net/http"

	"github.com/dhiliphari/tixr/internal/catalog"
)

func (app *application) listEventsHandler(w http.ResponseWriter, r *http.Request) {
	events, err := app.catalog.ListEvents(r.Context())
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if err := app.writeJSON(w, http.StatusOK, envelope{"events": events}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showEventHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	event, err := app.catalog.GetEvent(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, catalog.ErrNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if err := app.writeJSON(w, http.StatusOK, envelope{"event": event}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listSeatsHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// An unknown event should be a 404, not an empty list of seats.
	if _, err := app.catalog.GetEvent(r.Context(), id); err != nil {
		switch {
		case errors.Is(err, catalog.ErrNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	seats, err := app.booking.ListSeats(r.Context(), id)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if err := app.writeJSON(w, http.StatusOK, envelope{"seats": seats}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
