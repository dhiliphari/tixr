package booking

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound        = errors.New("booking: not found")
	ErrSeatUnavailable = errors.New("booking: seat unavailable")
)

const (
	SeatAvailable = "available"
	SeatBooked    = "booked"
)

type Seat struct {
	ID         int64  `json:"id"`
	EventID    int64  `json:"event_id"`
	Label      string `json:"label"`
	PriceCents int    `json:"price_cents"`
	Status     string `json:"status"`
}

type Booking struct {
	ID            int64     `json:"id"`
	EventID       int64     `json:"event_id"`
	SeatID        int64     `json:"seat_id"`
	CustomerEmail string    `json:"customer_email"`
	AmountCents   int       `json:"amount_cents"`
	CreatedAt     time.Time `json:"created_at"`
}

type Store struct {
	db *pgxpool.Pool
}

func (s *Store) ListSeats(ctx context.Context, eventID int64) ([]Seat, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, event_id, label, price_cents, status
		FROM seats
		WHERE event_id = $1
		ORDER BY id
	`, eventID)

	if err != nil {
		return nil, fmt.Errorf("list seats for event %d: %w", eventID, err)
	}

	seats, err := pgx.CollectRows(rows, pgx.RowToStructByPos[Seat])
	if err != nil {
		return nil, fmt.Errorf("list seats for event %d: %w", eventID, err)
	}

	return seats, nil
}

func (s *Store) Create(ctx context.Context, eventID, seatID int64, email string) (Booking, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Booking{}, fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback(ctx)

	var (
		status string
		price  int
	)
	err = tx.QueryRow(ctx, `
		SELECT status, price_cents
		FROM seats
		WHERE id = $1 AND event_id = $2
	`, seatID, eventID).Scan(&status, &price)

	if errors.Is(err, pgx.ErrNoRows) {
		return Booking{}, ErrNotFound
	}
	if err != nil {
		return Booking{}, fmt.Errorf("get seat %d: %w", seatID, err)
	}

	if status != SeatAvailable {
		return Booking{}, ErrSeatUnavailable
	}

	b := Booking{
		EventID:       eventID,
		SeatID:        seatID,
		CustomerEmail: email,
		AmountCents:   price,
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO bookings (event_id, seat_id, customer_email, amount_cents)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`, b.EventID, b.SeatID, b.CustomerEmail, b.AmountCents).Scan(&b.ID, &b.CreatedAt)

	if err != nil {
		return Booking{}, fmt.Errorf("insert booking: %w", err)
	}

	_, err = tx.Exec(ctx, `UPDATE seats SET status = $1 WHERE id = $2`, SeatBooked, seatID)
	if err != nil {
		return Booking{}, fmt.Errorf("mark seat %d booked: %w", seatID, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return Booking{}, fmt.Errorf("commit: %w", err)
	}

	return b, nil
}

func (s *Store) Get(ctx context.Context, id int64) (Booking, error) {
	var b Booking

	err := s.db.QueryRow(ctx, `
		SELECT id, event_id, seat_id, customer_email, amount_cents, created_at
		FROM bookings
		WHERE id = $1
	`, id).Scan(&b.ID, &b.EventID, &b.SeatID, &b.CustomerEmail, &b.AmountCents, &b.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return Booking{}, ErrNotFound
	}
	if err != nil {
		return Booking{}, fmt.Errorf("get booking %d: %w", id, err)
	}

	return b, nil
}
