package main

import (
	"fmt"
	"net/http"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if pv := recover(); pv != nil {
				w.Header().Add("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("panic: %v", pv))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
