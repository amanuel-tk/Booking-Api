package showtime

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	pgForeignKeyViolation = "23503"
	pgExclusionViolation  = "23P01"
	pgInvalidText         = "22P02"
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

func (r *Repository) CreateShowtime(ctx context.Context, st *Showtime) error {
	const q = `
		INSERT INTO showtimes (movie_id, screen_id, starts_at, ends_at, price_cents)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, is_active, created_at, updated_at`

	err := r.db.QueryRow(ctx, q,
		st.MovieID, st.ScreenID, st.StartsAt, st.EndsAt, st.PriceCents,
	).Scan(&st.ID, &st.IsActive, &st.CreatedAt, &st.UpdatedAt)
	if err != nil {
		switch pgCode(err) {
		case pgExclusionViolation:
			return ErrShowtimeOverlap
		case pgForeignKeyViolation, pgInvalidText:
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (r *Repository) GetShowtime(ctx context.Context, id string) (*Showtime, error) {
	const q = `
		SELECT id, movie_id, screen_id, starts_at, ends_at, price_cents,
		       is_active, created_at, updated_at
		FROM showtimes
		WHERE id = $1 AND is_active = true`

	var st Showtime
	err := r.db.QueryRow(ctx, q, id).Scan(
		&st.ID, &st.MovieID, &st.ScreenID, &st.StartsAt, &st.EndsAt, &st.PriceCents,
		&st.IsActive, &st.CreatedAt, &st.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || pgCode(err) == pgInvalidText {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &st, nil
}
