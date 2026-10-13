package movie

import "time"

type Movie struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	DurationMinutes int       `json:"duration_minutes"`
	Language        string    `json:"language"`
	Genres          []string  `json:"genres"`
	ReleaseDate     time.Time `json:"release_date"`
	PosterURL       string    `json:"poster_url"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
