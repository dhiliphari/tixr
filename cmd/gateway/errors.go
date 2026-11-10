package main

import (
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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

func (app *application) grpcErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	st, ok := status.FromError(err)
	if !ok {
		app.serverErrorResponse(w, r, err)
		return
	}

	switch st.Code() {
	case codes.NotFound:
		app.notFoundResponse(w, r)
	case codes.InvalidArgument:
		app.errorResponse(w, r, http.StatusBadRequest, st.Message())
	case codes.FailedPrecondition, codes.AlreadyExists:
		app.conflictResponse(w, r, st.Message())
	case codes.Unavailable:
		app.logError(r, err)
		app.errorResponse(w, r, http.StatusServiceUnavailable, "the service is temporarily unavailable, please try again later")
	case codes.DeadlineExceeded:
		app.logError(r, err)
		app.errorResponse(w, r, http.StatusGatewayTimeout, "the request took too long to process")
	default:
		app.serverErrorResponse(w, r, err)
	}
}
