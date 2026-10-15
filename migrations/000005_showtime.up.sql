CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE showtimes (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    movie_id    uuid        NOT NULL REFERENCES movies(id)  ON DELETE RESTRICT,
    screen_id   uuid        NOT NULL REFERENCES screens(id) ON DELETE RESTRICT,
    starts_at   timestamptz NOT NULL,
    ends_at     timestamptz NOT NULL,
    price_cents integer     NOT NULL,
    is_active   boolean     NOT NULL DEFAULT true,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT showtimes_time_order   CHECK (ends_at > starts_at),
    CONSTRAINT showtimes_price_nonneg CHECK (price_cents >= 0),

    CONSTRAINT showtimes_no_overlap EXCLUDE USING gist (
        screen_id WITH =,
        tstzrange(starts_at, ends_at, '[)') WITH &&
    ) WHERE (is_active)
);

CREATE INDEX idx_showtimes_movie_starts ON showtimes (movie_id, starts_at);