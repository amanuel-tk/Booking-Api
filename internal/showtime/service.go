package showtime

import (
	"context"
	"errors"
	"time"

	"github.com/amanuel-tk/Booking-Api/internal/movie"
	"github.com/amanuel-tk/Booking-Api/internal/theater"
)

const (
	CleaningBuffer = 15 * time.Minute
	MaxPriceCents  = 100_000_00
	dateLayout     = "2006-01-02"
)

type MovieGetter interface {
	GetMovie(ctx context.Context, id string) (*movie.Movie, error)
}

type ScreenGetter interface {
	GetScreenInfo(ctx context.Context, id string) (*theater.Screen, error)
}

type Repo interface {
	CreateShowtime(ctx context.Context, st *Showtime) error
	GetShowtime(ctx context.Context, id string) (*Showtime, error)
}

type Service struct {
	repo    Repo
	movies  MovieGetter
	screens ScreenGetter
	loc     *time.Location
	now     func() time.Time
}

func NewService(repo Repo, movies MovieGetter, screens ScreenGetter, loc *time.Location) *Service {
	return &Service{repo: repo, movies: movies, screens: screens, loc: loc, now: time.Now}
}

func (s *Service) CreateShowtime(ctx context.Context, req CreateShowtimeRequest) (*Showtime, error) {
	if req.MovieID == "" || req.ScreenID == "" {
		return nil, ValidationError{"movie_id and screen_id are required"}
	}
	if req.StartsAt.IsZero() {
		return nil, ValidationError{"starts_at is required (e.g. 2026-10-20T18:00:00+03:00)"}
	}
	if req.PriceCents < 0 || req.PriceCents > MaxPriceCents {
		return nil, ValidationError{"price_cents is out of range"}
	}
	if !req.StartsAt.After(s.now()) {
		return nil, ValidationError{"starts_at must be in the future"}
	}

	m, err := s.movies.GetMovie(ctx, req.MovieID)
	if err != nil {
		if errors.Is(err, movie.ErrNotFound) {
			return nil, ErrMovieNotFound
		}
		return nil, err
	}
	if !m.IsActive {
		return nil, ValidationError{"movie is not active"}
	}

	sc, err := s.screens.GetScreenInfo(ctx, req.ScreenID)
	if err != nil {
		if errors.Is(err, theater.ErrNotFound) {
			return nil, ErrScreenNotFound
		}
		return nil, err
	}
	_ = sc

	startsAt := req.StartsAt.UTC()
	st := &Showtime{
		MovieID:    req.MovieID,
		ScreenID:   req.ScreenID,
		StartsAt:   startsAt,
		EndsAt:     ComputeEndsAt(startsAt, m.DurationMinutes),
		PriceCents: req.PriceCents,
	}
	if err := s.repo.CreateShowtime(ctx, st); err != nil {
		return nil, err
	}
	return st, nil
}

func (s *Service) GetShowtime(ctx context.Context, id string) (*Showtime, error) {
	return s.repo.GetShowtime(ctx, id)
}

func ComputeEndsAt(start time.Time, durationMinutes int) time.Time {
	return start.Add(time.Duration(durationMinutes)*time.Minute + CleaningBuffer)
}
