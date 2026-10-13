package movie

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("movie not found")

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

const columns = `id, title, description, duration_minutes, language, genres,
	release_date, poster_url, is_active, created_at, updated_at`

func (r *Repository) ListMovies(ctx context.Context, search string, limit, offset int) ([]Movie, int, error) {

	var total int

	err := r.db.QueryRow(ctx, `
        SELECT COUNT(*)
        FROM movies
        WHERE is_active
          AND ($1 = '' OR title ILIKE '%' || $1 || '%')
    `, search).Scan(&total)

	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx, `SELECT `+columns+` FROM movies WHERE is_active AND ($1='' OR title ILIKE '%' || $1 || '%') ORDER BY release_date DESC, id LIMIT $2 OFFSET $3`, search, limit, offset)

	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	movies := []Movie{}

	for rows.Next() {
		var m Movie
		if err := rows.Scan(&m.ID, &m.Title, &m.Description, &m.DurationMinutes, &m.Language,
			&m.Genres, &m.ReleaseDate, &m.PosterURL, &m.IsActive, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, 0, err
		}
		movies = append(movies, m)
	}
	return movies, total, rows.Err()
}

func (r *Repository) GetMovie(ctx context.Context, id string) (*Movie, error) {
	var m Movie
	err := r.db.QueryRow(ctx, `SELECT `+columns+` FROM movies WHERE id=$1 AND is_active`, id).Scan(&m.ID, &m.Title, &m.Description, &m.DurationMinutes, &m.Language, &m.Genres, &m.ReleaseDate, &m.PosterURL, &m.IsActive, &m.CreatedAt, &m.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}
