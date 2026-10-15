package showtime

import (
	"errors"
	"time"
)

type Showtime struct {
	ID         string    `json:"id"`
	MovieID    string    `json:"movie_id"`
	ScreenID   string    `json:"screen_id"`
	StartsAt   time.Time `json:"starts_at"`
	EndsAt     time.Time `json:"ends_at"`
	PriceCents int       `json:"price_cents"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ListItem struct {
	ID          string    `json:"id"`
	StartsAt    time.Time `json:"starts_at"`
	EndsAt      time.Time `json:"ends_at"`
	PriceCents  int       `json:"price_cents"`
	ScreenID    string    `json:"screen_id"`
	ScreenName  string    `json:"screen_name"`
	TheaterID   string    `json:"theater_id"`
	TheaterName string    `json:"theater_name"`
}

type CreateShowtimeRequest struct {
	MovieID    string    `json:"movie_id"`
	ScreenID   string    `json:"screen_id"`
	StartsAt   time.Time `json:"starts_at"`
	PriceCents int       `json:"price_cents"`
}

var (
	ErrNotFound        = errors.New("showtime not found")
	ErrMovieNotFound   = errors.New("movie not found")
	ErrScreenNotFound  = errors.New("screen not found")
	ErrShowtimeOverlap = errors.New("another showtime overlaps on this screen")
)

type ValidationError struct{ Msg string }

func (e ValidationError) Error() string { return e.Msg }
