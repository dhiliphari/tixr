package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /v1/healthcheck", app.healthCheckHandler)

	mux.HandleFunc("GET /v1/events", app.listEventsHandler)
	mux.HandleFunc("GET /v1/events/{id}", app.showEventHandler)
	mux.HandleFunc("GET /v1/events/{id}/seats", app.listSeatsHandler)
	mux.HandleFunc("POST /v1/events/{id}/bookings", app.createBookingHandler)

	mux.HandleFunc("GET /v1/bookings/{id}", app.showBookingHandler)

	return mux
}
