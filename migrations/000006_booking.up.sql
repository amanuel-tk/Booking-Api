CREATE TABLE bookings (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      uuid        NOT NULL REFERENCES users(id)     ON DELETE RESTRICT,
    showtime_id  uuid        NOT NULL REFERENCES showtimes(id) ON DELETE RESTRICT,
    status       text        NOT NULL DEFAULT 'confirmed',
    total_cents  integer     NOT NULL CHECK (total_cents >= 0),
    cancelled_at timestamptz,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT bookings_status_valid    CHECK (status IN ('confirmed', 'cancelled')),
    CONSTRAINT bookings_id_showtime_key UNIQUE (id, showtime_id)
);

CREATE INDEX idx_bookings_user_created ON bookings (user_id, created_at DESC);

CREATE TABLE booking_seats (
    booking_id  uuid    NOT NULL,
    showtime_id uuid    NOT NULL,
    seat_id     uuid    NOT NULL REFERENCES seats(id) ON DELETE RESTRICT,
    price_cents integer NOT NULL CHECK (price_cents >= 0),
    is_active   boolean NOT NULL DEFAULT true,

    PRIMARY KEY (booking_id, seat_id),
    FOREIGN KEY (booking_id, showtime_id)
        REFERENCES bookings (id, showtime_id) ON DELETE RESTRICT
);

CREATE UNIQUE INDEX booking_seats_showtime_seat_active_key
    ON booking_seats (showtime_id, seat_id) WHERE is_active;