package main

import (
	"context"

	catalogv1 "github.com/dhiliphari/tixr/gen/catalog/v1"
	"github.com/dhiliphari/tixr/internal/catalog"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (app *application) ListEvents(ctx context.Context, req *catalogv1.ListEventsRequest) (*catalogv1.ListEventsResponse, error) {
	events, err := app.store.ListEvents(ctx)
	if err != nil {
		return nil, app.toStatus(err)
	}

	resp := &catalogv1.ListEventsResponse{
		Events: make([]*catalogv1.Event, 0, len(events)),
	}
	for _, e := range events {
		resp.Events = append(resp.Events, toProtoEvent(e))
	}

	return resp, nil
}

func (app *application) GetEvent(ctx context.Context, req *catalogv1.GetEventRequest) (*catalogv1.GetEventResponse, error) {
	e, err := app.store.GetEvent(ctx, req.GetId())
	if err != nil {
		return nil, app.toStatus(err)
	}

	return &catalogv1.GetEventResponse{Event: toProtoEvent(e)}, nil
}

func toProtoEvent(e catalog.Event) *catalogv1.Event {
	return &catalogv1.Event{
		Id:        e.ID,
		Name:      e.Name,
		Venue:     e.Venue,
		StartsAt:  timestamppb.New(e.StartsAt),
		CreatedAt: timestamppb.New(e.CreatedAt),
	}
}
