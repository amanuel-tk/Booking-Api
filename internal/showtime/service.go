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
	maxRangeDays   = 31
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
	ListByMovie(ctx context.Context, movieID string, from, to, after time.Time) ([]ListItem, error)
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

	_, err = s.screens.GetScreenInfo(ctx, req.ScreenID)
	if err != nil {
		if errors.Is(err, theater.ErrNotFound) {
			return nil, ErrScreenNotFound
		}
		return nil, err
	}

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

func (s *Service) ListByMovie(ctx context.Context, movieID, fromStr, toStr string) ([]ListItem, error) {
	now := s.now()
	if fromStr == "" {
		fromStr = now.In(s.loc).Format(dateLayout)
	}
	if toStr == "" {
		toStr = fromStr
	}

	from, _, err := dayRange(fromStr, s.loc)
	if err != nil {
		return nil, err
	}
	_, to, err := dayRange(toStr, s.loc)
	if err != nil {
		return nil, err
	}

	if !to.After(from) {
		return nil, ValidationError{"to must not be before from"}
	}
	if to.Sub(from) > maxRangeDays*24*time.Hour {
		return nil, ValidationError{"range must be at most 31 days"}
	}

	return s.repo.ListByMovie(ctx, movieID, from, to, now)
}

func dayRange(date string, loc *time.Location) (time.Time, time.Time, error) {
	day, err := time.ParseInLocation(dateLayout, date, loc)
	if err != nil {
		return time.Time{}, time.Time{}, ValidationError{"date must be YYYY-MM-DD"}
	}
	return day, day.AddDate(0, 0, 1), nil
}

func ComputeEndsAt(start time.Time, durationMinutes int) time.Time {
	return start.Add(time.Duration(durationMinutes)*time.Minute + CleaningBuffer)
}
