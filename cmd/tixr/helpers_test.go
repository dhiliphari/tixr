package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	t.Parallel()

	app := application{}
	rr := httptest.NewRecorder()
	headers := http.Header{"X-Request-Id": []string{"abc123"}}

	err := app.writeJSON(rr, http.StatusCreated, envelope{"hello": "world"}, headers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	res := rr.Result()

	if res.StatusCode != http.StatusCreated {
		t.Errorf("status: got %d, want %d", res.StatusCode, http.StatusCreated)
	}
	if got := res.Header.Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type: got %q, want %q", got, "application/json")
	}
	if got := res.Header.Get("X-Request-Id"); got != "abc123" {
		t.Errorf("X-Request-Id: got %q, want %q", got, "abc123")
	}

	var body map[string]string
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["hello"] != "world" {
		t.Errorf("body[hello]: got %q, want %q", body["hello"], "world")
	}
}

func TestWriteJSONMarshalErrorWritesNothing(t *testing.T) {
	t.Parallel()

	app := &application{}
	rr := httptest.NewRecorder()

	// Channels can't be marshalled to JSON.
	err := app.writeJSON(rr, http.StatusOK, envelope{"bad": make(chan int)}, nil)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}

	if rr.Body.Len() != 0 {
		t.Errorf("body should be empty, got %q", rr.Body.String())
	}
	if rr.Header().Get("Content-Type") != "" {
		t.Errorf("Content-Type should not be set, got %q", rr.Header().Get("Content-Type"))
	}
}
