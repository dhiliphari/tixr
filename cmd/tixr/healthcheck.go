package main

import (
	"context"
	"net/http"
	"time"
)

func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	status := "available"
	dbStatus := "ok"
	code := http.StatusOK

	if err := app.db.Ping(ctx); err != nil {
		app.logger.Warn("healthcheck: database ping failed", "error", err)
		status = "degraded"
		dbStatus = "unavailable"
		code = http.StatusServiceUnavailable
	}

	env := envelope{
		"status": status,
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     version,
		},
		"checks": map[string]string{
			"database": dbStatus,
		},
	}

	if err := app.writeJSON(w, code, env, nil); err != nil {
		app.logger.Error("write healthcheck response", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
