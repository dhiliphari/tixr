CREATE TABLE IF NOT EXISTS events (
    id         bigserial   PRIMARY KEY,
    name       text        NOT NULL,
    venue      text        NOT NULL,
    starts_at  timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);