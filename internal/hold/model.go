package hold

import (
	"errors"
	"time"
)

type Request struct {
	SeatIDs []string `json:"seat_ids" minItems:"1" maxItems:"6"`
}

type Result struct {
	ShowtimeID string    `json:"showtime_id" format:"uuid"`
	SeatIDs    []string  `json:"seat_ids"`
	ExpiresAt  time.Time `json:"expires_at"`
}

type SeatsUnavailableResponse struct {
	Error   string   `json:"error" example:"some seats are not available"`
	SeatIDs []string `json:"seat_ids"`
}

var ErrNotFound = errors.New("showtime not found")

type ValidationError struct{ Msg string }

func (e ValidationError) Error() string { return e.Msg }

type SeatsUnavailableError struct{ SeatIDs []string }

func (e SeatsUnavailableError) Error() string { return "some seats are not available" }
