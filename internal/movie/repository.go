package movie

import (
	"context"

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

func (r *Repository) Create(ctx context.Context, m *Movie) error {

	return r.db.QueryRow(ctx, `INSERT INTO movies (title, description,duration_minutes,language,genres, release_date,poster_url) VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id,is_active,created_at,updated_at`, m.Title, m.Description, m.DurationMinutes, m.Language, m.Genres, m.ReleaseDate, m.PosterURL).Scan(&m.ID, &m.IsActive, &m.CreatedAt, &m.UpdatedAt)
}
