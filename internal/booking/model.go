package booking

import (
	"errors"
	"time"
)

const (
	StatusConfirmed = "confirmed"
	StatusCancelled = "cancelled"
)

type SeatStatus string

const (
	SeatFree   SeatStatus = "free"
	SeatHeld   SeatStatus = "held"
	SeatBooked SeatStatus = "booked"
)

type CreateRequest struct {
	SeatIDs []string `json:"seat_ids" minItems:"1" maxItems:"6"`
}

type NewBooking struct {
	ID         string
	UserID     string
	ShowtimeID string
	SeatIDs    []string
	PriceCents int
	TotalCents int
}

type BookingView struct {
	ID          string     `json:"id"`
	Status      string     `json:"status"`
	TotalCents  int        `json:"total_cents"`
	Seats       []string   `json:"seats"`
	ShowtimeID  string     `json:"showtime_id"`
	StartsAt    time.Time  `json:"starts_at"`
	MovieTitle  string     `json:"movie_title"`
	TheaterName string     `json:"theater_name"`
	ScreenName  string     `json:"screen_name"`
	CreatedAt   time.Time  `json:"created_at"`
	CancelledAt *time.Time `json:"cancelled_at,omitempty"`
}

type SeatView struct {
	ID     string     `json:"id" format:"uuid"`
	Number int        `json:"number" example:"1"`
	Status SeatStatus `json:"status" enums:"free,held,booked"`
}

type SeatMapRow struct {
	Row   string     `json:"row"`
	Seats []SeatView `json:"seats"`
}

type SeatMap struct {
	ShowtimeID string       `json:"showtime_id"`
	ScreenID   string       `json:"screen_id"`
	Available  int          `json:"available"`
	Rows       []SeatMapRow `json:"rows"`
}

var (
	ErrNotFound         = errors.New("booking not found")
	ErrShowtimeNotFound = errors.New("showtime not found")
	ErrHoldNotOwned     = errors.New("your hold has expired or does not belong to you")
	ErrSeatsTaken       = errors.New("one or more seats are already booked")
	ErrAlreadyCancelled = errors.New("booking is already cancelled")
	ErrTooLateToCancel  = errors.New("bookings cannot be cancelled within 2 hours of the showtime")
)

type ValidationError struct{ Msg string }

func (e ValidationError) Error() string { return e.Msg }
