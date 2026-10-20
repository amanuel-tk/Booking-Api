package movie

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}
type CreateRequest struct {
	Title           string   `json:"title" example:"Inception" maxLength:"200"`
	Description     string   `json:"description" example:"A science fiction thriller"`
	DurationMinutes int      `json:"duration_minutes" example:"148" minimum:"1" maximum:"600"`
	Language        string   `json:"language" example:"English"`
	Genres          []string `json:"genres"`
	ReleaseDate     string   `json:"release_date" example:"2010-07-16" format:"date"`
	PosterURL       string   `json:"poster_url" example:"https://example.com/poster.jpg"`
}

type ValidationError struct {
	Msg string
}

func (e *ValidationError) Error() string {
	return e.Msg
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (*Movie, error) {
	title := strings.TrimSpace(strings.ToLower(req.Title))
	description := strings.TrimSpace(strings.ToLower(req.Description))
	language := strings.TrimSpace(strings.ToLower(req.Language))
	posterUrl := strings.TrimSpace(strings.ToLower(req.PosterURL))

	if title == "" || len(title) > 200 {
		return nil, &ValidationError{"title is required (max 200 characters)"}
	}

	if req.DurationMinutes <= 0 || req.DurationMinutes > 600 {
		return nil, &ValidationError{"duration_minutes must be between 1 and 600"}
	}

	released, err := time.Parse("2006-01-02", req.ReleaseDate)

	if err != nil {
		return nil, &ValidationError{"released date must be in YYYY-MM-DD format"}
	}

	genres := make([]string, 0, len(req.Genres))

	for _, g := range req.Genres {
		if g = strings.TrimSpace(strings.ToLower(g)); g != "" {
			genres = append(genres, g)
		}
	}

	m := &Movie{
		Title:           title,
		Description:     description,
		DurationMinutes: req.DurationMinutes,
		Language:        language,
		Genres:          genres,
		ReleaseDate:     released,
		PosterURL:       posterUrl,
	}

	if err := s.repo.Create(ctx, m); err != nil {
		return nil, err
	}

	return m, nil

}

func (s *Service) ListMovies(ctx context.Context, search string, limit, page int) ([]Movie, int, error) {

	if limit < 1 || limit > 100 {
		limit = 20
	}
	if page < 1 {
		page = 1
	}
	return s.repo.ListMovies(ctx, search, limit, (page-1)*limit)
}

func (s *Service) GetMovie(ctx context.Context, id string) (*Movie, error) {

	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrNotFound
	}

	return s.repo.GetMovie(ctx, id)
}
