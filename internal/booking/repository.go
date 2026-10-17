package booking

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	pgUniqueViolation = "23505"
	pgInvalidText     = "22P02"
	maxListRows       = 100
)

func pgCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func mapWriteErr(err error, what string) error {
	if pgCode(err) == pgUniqueViolation {
		return ErrSeatsTaken
	}
	return fmt.Errorf("%s: %w", what, err)
}
func (r *Repository) CreateBooking(ctx context.Context, nb *NewBooking) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx,
		`INSERT INTO bookings (user_id, showtime_id, total_cents)
		 VALUES ($1, $2, $3)
		 RETURNING id`,
		nb.UserID, nb.ShowtimeID, nb.TotalCents,
	).Scan(&nb.ID)
	if err != nil {
		return fmt.Errorf("insert booking: %w", err)
	}

	rows := make([][]any, len(nb.SeatIDs))
	for i, seatID := range nb.SeatIDs {
		rows[i] = []any{nb.ID, nb.ShowtimeID, seatID, nb.PriceCents}
	}

	n, err := tx.CopyFrom(
		ctx,
		pgx.Identifier{"booking_seats"},
		[]string{"booking_id", "showtime_id", "seat_id", "price_cents"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return mapWriteErr(err, "copy booking seats")
	}
	if int(n) != len(nb.SeatIDs) {
		return fmt.Errorf("copy booking seats: inserted %d of %d", n, len(nb.SeatIDs))
	}

	if err := tx.Commit(ctx); err != nil {
		return mapWriteErr(err, "commit")
	}
	return nil
}

func (r *Repository) BookedAmong(ctx context.Context, showtimeID string, seatIDs []string) ([]string, error) {
	const q = `
		SELECT seat_id::text
		FROM booking_seats
		WHERE showtime_id = $1 AND is_active AND seat_id::text = ANY($2)`

	rows, err := r.db.Query(ctx, q, showtimeID, seatIDs)
	if err != nil {
		return nil, fmt.Errorf("booked among: %w", err)
	}
	defer rows.Close()

	booked := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan booked seat: %w", err)
		}
		booked = append(booked, id)
	}
	return booked, rows.Err()
}

func (r *Repository) BookedSeatIDs(ctx context.Context, showtimeID string) (map[string]bool, error) {
	const q = `SELECT seat_id::text FROM booking_seats WHERE showtime_id = $1 AND is_active`

	rows, err := r.db.Query(ctx, q, showtimeID)
	if err != nil {
		return nil, fmt.Errorf("booked seat ids: %w", err)
	}
	defer rows.Close()

	booked := make(map[string]bool)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan booked seat id: %w", err)
		}
		booked[id] = true
	}
	return booked, rows.Err()
}

const viewSelect = `
	SELECT b.id, b.status, b.total_cents, b.created_at, b.cancelled_at,
	       st.id, st.starts_at, m.title, t.name, sc.name,
	       array_agg(s.row_label || s.seat_number::text
	                 ORDER BY s.row_label, s.seat_number)
	FROM bookings b
	JOIN showtimes st     ON st.id = b.showtime_id
	JOIN movies m         ON m.id  = st.movie_id
	JOIN screens sc       ON sc.id = st.screen_id
	JOIN theaters t       ON t.id  = sc.theater_id
	JOIN booking_seats bs ON bs.booking_id = b.id
	JOIN seats s          ON s.id  = bs.seat_id
	WHERE b.user_id = $1`

const viewGroup = ` GROUP BY b.id, st.id, m.id, t.id, sc.id`

type scanner interface {
	Scan(dest ...any) error
}

func scanView(row scanner) (BookingView, error) {
	var v BookingView
	err := row.Scan(
		&v.ID, &v.Status, &v.TotalCents, &v.CreatedAt, &v.CancelledAt,
		&v.ShowtimeID, &v.StartsAt, &v.MovieTitle, &v.TheaterName, &v.ScreenName,
		&v.Seats,
	)
	return v, err
}

func (r *Repository) GetForUser(ctx context.Context, bookingID, userID string) (*BookingView, error) {
	q := viewSelect + ` AND b.id = $2` + viewGroup

	v, err := scanView(r.db.QueryRow(ctx, q, userID, bookingID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || pgCode(err) == pgInvalidText {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get booking: %w", err)
	}
	return &v, nil
}

func (r *Repository) ListForUser(ctx context.Context, userID string) ([]BookingView, error) {
	q := viewSelect + viewGroup + fmt.Sprintf(` ORDER BY b.created_at DESC LIMIT %d`, maxListRows)

	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("list bookings: %w", err)
	}
	defer rows.Close()

	out := []BookingView{}
	for rows.Next() {
		v, err := scanView(rows)
		if err != nil {
			return nil, fmt.Errorf("scan booking: %w", err)
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (r *Repository) CancelBooking(ctx context.Context, bookingID, userID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx,
		`UPDATE bookings
		 SET status = 'cancelled', cancelled_at = now(), updated_at = now()
		 WHERE id = $1 AND user_id = $2 AND status = 'confirmed'`,
		bookingID, userID,
	)
	if err != nil {
		if pgCode(err) == pgInvalidText {
			return ErrNotFound
		}
		return fmt.Errorf("cancel booking: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrAlreadyCancelled
	}

	if _, err := tx.Exec(ctx,
		`UPDATE booking_seats SET is_active = false WHERE booking_id = $1`, bookingID,
	); err != nil {
		return fmt.Errorf("free seats: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}
