package theater

import (
	"errors"
	"time"
)

type Theater struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	City      string    `json:"city"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Screen struct {
	ID          string    `json:"id"`
	TheaterID   string    `json:"theater_id"`
	Name        string    `json:"name"`
	RowCount    int       `json:"row_count"`
	SeatsPerRow int       `json:"seats_per_row"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Seat struct {
	ID         string `json:"id,omitempty" format:"uuid"`
	ScreenID   string `json:"-"`
	RowLabel   string `json:"row_label" example:"A"`
	SeatNumber int    `json:"seat_number" example:"1"`
}

type SeatRow struct {
	Row   string `json:"row" example:"A"`
	Seats []Seat `json:"seats"`
}

type ScreenLayout struct {
	Screen Screen    `json:"screen"`
	Rows   []SeatRow `json:"rows"`
}

var (
	ErrNotFound        = errors.New("not found")
	ErrScreenNameTaken = errors.New("screen name already used in this theater")
	ErrTheaterTaken    = errors.New("theater already exists in this city")
)

type ValidationError struct{ Msg string }

func (e ValidationError) Error() string { return e.Msg }
