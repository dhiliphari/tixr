CREATE TABLE IF NOT EXISTS seats (
    id          bigserial PRIMARY KEY,
    event_id    bigint    NOT NULL REFERENCES events (id) ON DELETE CASCADE,
    label       text      NOT NULL,
    price_cents integer   NOT NULL CHECK (price_cents >= 0),
    status      text      NOT NULL DEFAULT 'available'
                          CHECK (status IN ('available', 'booked')),
    version     integer   NOT NULL DEFAULT 1,
    UNIQUE (event_id, label)
);