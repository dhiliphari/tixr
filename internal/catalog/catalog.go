package catalog

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("catalog: not found")

type Event struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Venue     string    `json:"venue"`
	StartsAt  time.Time `json:"starts_at"`
	CreatedAt time.Time `json:"created_at"`
}

type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) ListEvents(ctx context.Context) ([]Event, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, name, venue, starts_at, created_at
		FROM events
		ORDER BY starts_at, id
	`)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}

	events, err := pgx.CollectRows(rows, pgx.RowToStructByPos[Event])
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}

	return events, nil
}

func (s *Store) GetEvent(ctx context.Context, id int64) (Event, error) {
	var e Event

	err := s.db.QueryRow(ctx, `
		SELECT id, name, venue, starts_at, created_at
		FROM events
		WHERE id = $1
	`, id).Scan(&e.ID, &e.Name, &e.Venue, &e.StartsAt, &e.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return Event{}, ErrNotFound
	}
	if err != nil {
		return Event{}, fmt.Errorf("get event %d: %w", id, err)
	}

	return e, nil
}
