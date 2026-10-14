package theater

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

const (
	pgUniqueViolation     = "23505"
	pgForeignKeyViolation = "23503"
	pgInvalidText         = "22P02"
)

func pgCode(err error) string {
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		return pgErr.Code
	}
	return ""
}

func (r *Repository) CreateTheater(ctx context.Context, t *Theater) error {
	err := r.db.QueryRow(ctx, "INSERT INTO theaters (name,city) VALUES ($1,$2) RETURNING id,is_active,created_at,updated_at", t.Name, t.City).Scan(&t.ID, &t.IsActive, &t.CreatedAt, &t.UpdatedAt)

	if err != nil {
		if pgCode(err) == pgUniqueViolation {
			return ErrTheaterTaken
		}
		return err
	}
	return nil
}

func (r *Repository) CreateScreenWithSeats(ctx context.Context, s *Screen, seats []Seat) error {
	tx, err := r.db.Begin(ctx)

	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, "INSERT INTO screens (theater_id,name,row_count,seats_per_row) VALUES ($1,$2,$3,$4) RETURNING id,is_active,created_at,updated_at", s.TheaterID, s.Name, s.RowCount, s.SeatsPerRow).Scan(&s.ID, &s.IsActive, &s.CreatedAt, &s.UpdatedAt)

	if err != nil {
		switch pgCode(err) {
		case pgUniqueViolation:
			return ErrScreenNameTaken
		case pgForeignKeyViolation, pgInvalidText:
			return ErrNotFound
		}
		return err
	}

	rows := make([][]any, len(seats))

	for i, seat := range seats {
		rows[i] = []any{s.ID, seat.RowLabel, seat.SeatNumber}
	}

	n, err := tx.CopyFrom(ctx, pgx.Identifier{"seats"}, []string{"screen_id", "row_label", "seat_number"}, pgx.CopyFromRows(rows))

	if err != nil {
		return fmt.Errorf("seat copy:%w", err)
	}
	if int(n) != len(seats) {
		return fmt.Errorf("copy seats: inserted %d of %d", n, len(seats))
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil

}

func (r *Repository) GetTheaters(ctx context.Context) ([]Theater, error) {
	theaters := []Theater{}

	rows, err := r.db.Query(ctx, "SELECT id,name,city,is_active,created_at,updated_at FROM theaters where is_active ORDER BY created_at")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var t Theater

		if err := rows.Scan(&t.ID, &t.Name, &t.City, &t.IsActive, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		theaters = append(theaters, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return theaters, nil
}
