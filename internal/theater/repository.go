package theater

import (
	"context"
	"errors"

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
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
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
