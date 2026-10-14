CREATE TABLE theaters (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name       text        NOT NULL,
    city       text        NOT NULL,
    is_active  boolean     NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (name, city)
);

CREATE TABLE screens (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    theater_id    uuid        NOT NULL REFERENCES theaters(id) ON DELETE RESTRICT,
    name          text        NOT NULL,
    row_count     int         NOT NULL CHECK (row_count BETWEEN 1 AND 26),
    seats_per_row int         NOT NULL CHECK (seats_per_row BETWEEN 1 AND 50),
    is_active     boolean     NOT NULL DEFAULT true,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    UNIQUE (theater_id, name)
);

CREATE TABLE seats (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    screen_id   uuid NOT NULL REFERENCES screens(id) ON DELETE RESTRICT,
    row_label   text NOT NULL,
    seat_number int  NOT NULL CHECK (seat_number > 0),
    UNIQUE (screen_id, row_label, seat_number)
);