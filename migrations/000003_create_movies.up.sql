CREATE TABLE movies (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    title            text NOT NULL,
    description      text NOT NULL DEFAULT '',
    duration_minutes int  NOT NULL CHECK (duration_minutes > 0),
    language         text NOT NULL DEFAULT '',
    genres           text[] NOT NULL DEFAULT '{}',
    release_date     date NOT NULL,
    poster_url       text NOT NULL DEFAULT '',
    is_active        boolean NOT NULL DEFAULT true,
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX movies_release_date_idx ON movies(release_date DESC);