package showtime

import (
	"context"
	"errors"
	"time"

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

func (r *Repository) ListByMovie(ctx context.Context, movieID string, from, to, after time.Time) ([]ListItem, error) {
	const q = `
		SELECT st.id, st.starts_at, st.ends_at, st.price_cents,
		       sc.id, sc.name, t.id, t.name
		FROM showtimes st
		JOIN screens  sc ON sc.id = st.screen_id
		JOIN theaters t  ON t.id  = sc.theater_id
		WHERE st.movie_id = $1
		  AND st.is_active = true
		  AND st.starts_at >= $2
		  AND st.starts_at <  $3
		  AND st.starts_at >  $4
		ORDER BY st.starts_at`

	rows, err := r.db.Query(ctx, q, movieID, from, to, after)
	if err != nil {
		if pgCode(err) == pgInvalidText {
			return []ListItem{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	items := []ListItem{}
	for rows.Next() {
		var it ListItem
		if err := rows.Scan(
			&it.ID, &it.StartsAt, &it.EndsAt, &it.PriceCents,
			&it.ScreenID, &it.ScreenName, &it.TheaterID, &it.TheaterName,
		); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
