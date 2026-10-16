package hold

import (
	"errors"
	"time"
)

type Request struct {
	SeatIDs []string `json:"seat_ids"`
}

type Result struct {
	ShowtimeID string    `json:"showtime_id"`
	SeatIDs    []string  `json:"seat_ids"`
	ExpiresAt  time.Time `json:"expires_at"`
}

var ErrNotFound = errors.New("showtime not found")

type ValidationError struct{ Msg string }

func (e ValidationError) Error() string { return e.Msg }

// SeatsUnavailableError carries WHICH seats were taken, so the API can show them.
type SeatsUnavailableError struct{ SeatIDs []string }

func (e SeatsUnavailableError) Error() string { return "some seats are not available" }
