CREATE TABLE IF NOT EXISTS bookings (
    id             bigserial   PRIMARY KEY,
    event_id       bigint      NOT NULL REFERENCES events (id),
    seat_id        bigint      NOT NULL REFERENCES seats (id),
    customer_email citext      NOT NULL,
    amount_cents   integer     NOT NULL,
    created_at     timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS bookings_event_id_idx ON bookings (event_id);