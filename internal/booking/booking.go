package booking

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound        = errors.New("booking: not found")
	ErrSeatUnavailable = errors.New("booking: seat unavailable")
	ErrSoldOut         = errors.New("booking: sold out")
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

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{
		db: db,
	}
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

	// Check and claim in one statement. If another transaction is claiming
	// the same seat, this waits for it, then re-evaluates the WHERE clause
	// against the seat's new state.
	var price int
	err = tx.QueryRow(ctx, `
		UPDATE seats
		SET status = $1, version = version + 1
		WHERE id = $2 AND event_id = $3 AND status = $4
		RETURNING price_cents`,
		SeatBooked, seatID, eventID, SeatAvailable).Scan(&price)
	if errors.Is(err, pgx.ErrNoRows) {
		return Booking{}, whyNotClaimed(ctx, tx, eventID, seatID)
	}
	if err != nil {
		return Booking{}, fmt.Errorf("claim seat %d: %w", seatID, err)
	}

	b, err := insertBooking(ctx, tx, eventID, seatID, email, price)
	if err != nil {
		return Booking{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return Booking{}, fmt.Errorf("commit: %w", err)
	}

	return b, nil
}

// CreateBestAvailable books the lowest-numbered available seat.
func (s *Store) CreateBestAvailable(ctx context.Context, eventID int64, email string) (Booking, error) {
	var seatID int64
	err := s.db.QueryRow(ctx, `
		SELECT id
		FROM seats
		WHERE event_id = $1 AND status = $2
		ORDER BY id
		LIMIT 1`, eventID, SeatAvailable).Scan(&seatID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Booking{}, ErrSoldOut
	}
	if err != nil {
		return Booking{}, fmt.Errorf("find available seat: %w", err)
	}

	return s.Create(ctx, eventID, seatID, email)
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

// whyNotClaimed works out why a conditional update matched no rows: either
// the seat doesn't exist for this event, or it isn't available.
func whyNotClaimed(ctx context.Context, tx pgx.Tx, eventID, seatID int64) error {
	var exists bool
	err := tx.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM seats WHERE id = $1 AND event_id = $2)`,
		seatID, eventID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check seat %d: %w", seatID, err)
	}

	if !exists {
		return ErrNotFound
	}
	return ErrSeatUnavailable
}

// uniqueViolation is Postgres's SQLSTATE for a unique constraint violation.
const uniqueViolation = "23505"

func insertBooking(ctx context.Context, tx pgx.Tx, eventID, seatID int64, email string, price int) (Booking, error) {
	b := Booking{
		EventID:       eventID,
		SeatID:        seatID,
		CustomerEmail: email,
		AmountCents:   price,
	}

	err := tx.QueryRow(ctx, `
		INSERT INTO bookings (event_id, seat_id, customer_email, amount_cents)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`,
		b.EventID, b.SeatID, b.CustomerEmail, b.AmountCents).Scan(&b.ID, &b.CreatedAt)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
		return Booking{}, ErrSeatUnavailable
	}
	if err != nil {
		return Booking{}, fmt.Errorf("insert booking: %w", err)
	}

	return b, nil
}
