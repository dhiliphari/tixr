//go:build integration

package booking_test

import (
	"context"
	"errors"
	"testing"

	"github.com/dhiliphari/tixr/internal/booking"
	"github.com/dhiliphari/tixr/internal/testdb"
	"github.com/jackc/pgx/v5/pgxpool"
)

// seedSeat creates an event with a single seat priced at 5000 cents.
func seedSeat(t *testing.T, db *pgxpool.Pool) (eventID, seatID int64) {
	t.Helper()
	ctx := context.Background()

	err := db.QueryRow(ctx, `
		INSERT INTO events (name, venue, starts_at)
		VALUES ('Test Event', 'Test Hall', now() + interval '1 day')
		RETURNING id`).Scan(&eventID)
	if err != nil {
		t.Fatalf("insert event: %v", err)
	}

	err = db.QueryRow(ctx, `
		INSERT INTO seats (event_id, label, price_cents)
		VALUES ($1, 'A1', 5000)
		RETURNING id`, eventID).Scan(&seatID)
	if err != nil {
		t.Fatalf("insert seat: %v", err)
	}

	return eventID, seatID
}

func TestStoreCreate(t *testing.T) {
	db := testdb.New(t)
	store := booking.NewStore(db)
	ctx := context.Background()

	eventID, seatID := seedSeat(t, db)

	b, err := store.Create(ctx, eventID, seatID, "Ada@Example.com")
	if err != nil {
		t.Fatalf("first booking: %v", err)
	}
	if b.AmountCents != 5000 {
		t.Errorf("amount: got %d, want 5000", b.AmountCents)
	}

	_, err = store.Create(ctx, eventID, seatID, "grace@example.com")
	if !errors.Is(err, booking.ErrSeatUnavailable) {
		t.Errorf("second booking: got error %v, want ErrSeatUnavailable", err)
	}

	_, err = store.Create(ctx, eventID, 999_999, "grace@example.com")
	if !errors.Is(err, booking.ErrNotFound) {
		t.Errorf("unknown seat: got error %v, want ErrNotFound", err)
	}

	got, err := store.Get(ctx, b.ID)
	if err != nil {
		t.Fatalf("get booking: %v", err)
	}
	if got.CustomerEmail != "Ada@Example.com" {
		t.Errorf("email: got %q, want %q", got.CustomerEmail, "Ada@Example.com")
	}

	var n int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM bookings WHERE seat_id = $1`, seatID).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("bookings for seat: got %d, want 1", n)
	}
}
