package main

import (
	"net/http"

	bookingv1 "github.com/dhiliphari/tixr/gen/booking/v1"
	catalogv1 "github.com/dhiliphari/tixr/gen/catalog/v1"
)

func (app *application) listEventsHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := app.catalog.ListEvents(r.Context(), &catalogv1.ListEventsRequest{})
	if err != nil {
		app.grpcErrorResponse(w, r, err)
		return
	}

	events := make([]event, 0, len(resp.Events))
	for _, e := range resp.Events {
		events = append(events, eventFromProto(e))
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

	resp, err := app.catalog.GetEvent(r.Context(), &catalogv1.GetEventRequest{Id: id})
	if err != nil {
		app.grpcErrorResponse(w, r, err)
		return
	}

	event := eventFromProto(resp.Event)

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
	if _, err := app.catalog.GetEvent(r.Context(), &catalogv1.GetEventRequest{Id: id}); err != nil {
		app.grpcErrorResponse(w, r, err)
		return
	}

	resp, err := app.booking.ListSeats(r.Context(), &bookingv1.ListSeatsRequest{EventId: id})
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	seats := make([]seat, 0, len(resp.Seats))
	for _, s := range resp.Seats {
		seats = append(seats, seatFromProto(s))
	}

	if err := app.writeJSON(w, http.StatusOK, envelope{"seats": seats}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
