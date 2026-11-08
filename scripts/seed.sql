-- Resets all Tixr data to a known state. DELETES EVERYTHING.
TRUNCATE bookings, seats, events RESTART IDENTITY CASCADE;

INSERT INTO events (name, venue, starts_at) VALUES
    ('Gophers Live',        'Main Hall', now() + interval '30 days'),
    ('The Flash Sale Tour', 'Arena',     now() + interval '60 days');

-- Event 1: 20 seats (rows A–B, seats 1–10). Seat ids 1–20.
INSERT INTO seats (event_id, label, price_cents)
SELECT 1, chr(64 + r) || s, 5000
FROM generate_series(1, 2) AS r, generate_series(1, 10) AS s
ORDER BY r, s;

-- Event 2: 100 seats (rows A–J). Seat ids 21–120. Rows A and B are premium.
INSERT INTO seats (event_id, label, price_cents)
SELECT 2, chr(64 + r) || s, CASE WHEN r <= 2 THEN 15000 ELSE 8000 END
FROM generate_series(1, 10) AS r, generate_series(1, 10) AS s
ORDER BY r, s;