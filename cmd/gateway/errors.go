package main

import "net/http"

func (app *application) logError(r *http.Request, err error) {
	app.logger.Error("request failed", "error", err, "method", r.Method, "uri", r.URL.RequestURI())
}

func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
	if err := app.writeJSON(w, status, envelope{"error": message}, nil); err != nil {
		app.logError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)
	app.errorResponse(w, r, http.StatusInternalServerError, "the server encountered a problem and could not process your request")
}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, http.StatusNotFound, "the requested resource could not be found")
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errs map[string]string) {
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errs)
}

func (app *application) conflictResponse(w http.ResponseWriter, r *http.Request, message string) {
	app.errorResponse(w, r, http.StatusConflict, message)
}
