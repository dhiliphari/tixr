package main

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGRPCErrorResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want int
	}{
		{"not found", status.Error(codes.NotFound, "event not found"), http.StatusNotFound},
		{"invalid argument", status.Error(codes.InvalidArgument, "bad email"), http.StatusBadRequest},
		{"failed precondition", status.Error(codes.FailedPrecondition, "seat taken"), http.StatusConflict},
		{"unavailable", status.Error(codes.Unavailable, "dial tcp 10.0.0.7:50051"), http.StatusServiceUnavailable},
		{"deadline exceeded", status.Error(codes.DeadlineExceeded, "slow"), http.StatusGatewayTimeout},
		{"internal", status.Error(codes.Internal, "internal error"), http.StatusInternalServerError},
		{"not a status error", errors.New("plain error"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app := &application{logger: slog.New(slog.NewJSONHandler(io.Discard, nil))}
			rr := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)

			app.grpcErrorResponse(rr, r, tt.err)

			if rr.Code != tt.want {
				t.Errorf("status: got %d, want %d", rr.Code, tt.want)
			}
			if strings.Contains(rr.Body.String(), "10.0.0.7") {
				t.Errorf("internal address leaked into the response: %s", rr.Body.String())
			}
		})
	}
}
