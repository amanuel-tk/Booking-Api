package hold

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/amanuel-tk/Booking-Api/internal/showtime"
)

const MaxSeatsPerHold = 6

// Small interfaces, defined where they are used.
type ShowtimeGetter interface {
	GetShowtime(ctx context.Context, id string) (*showtime.Showtime, error)
}

type SeatChecker interface {
	SeatsBelongToScreen(ctx context.Context, screenID string, seatIDs []string) (bool, error)
}

// BookedChecker returns which of the given seats are already booked.
// booking.Repository satisfies it, and we never import the booking package.
type BookedChecker interface {
	BookedAmong(ctx context.Context, showtimeID string, seatIDs []string) ([]string, error)
}

type SeatHolder interface {
	Hold(ctx context.Context, showtimeID string, seatIDs []string, userID string) ([]string, error)
}

type Service struct {
	holder    SeatHolder
	showtimes ShowtimeGetter
	seats     SeatChecker
	booked    BookedChecker
	ttl       time.Duration
	now       func() time.Time
}

func NewService(holder SeatHolder, showtimes ShowtimeGetter, seats SeatChecker, booked BookedChecker, ttl time.Duration) *Service {
	return &Service{holder: holder, showtimes: showtimes, seats: seats, booked: booked, ttl: ttl, now: time.Now}
}

func (s *Service) HoldSeats(ctx context.Context, showtimeID, userID string, seatIDs []string) (*Result, error) {
	seatIDs, err := cleanSeatIDs(seatIDs)
	if err != nil {
		return nil, err
	}

	st, err := s.showtimes.GetShowtime(ctx, showtimeID)
	if err != nil {
		if errors.Is(err, showtime.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if !st.StartsAt.After(s.now()) {
		return nil, ValidationError{"showtime has already started"}
	}

	ok, err := s.seats.SeatsBelongToScreen(ctx, st.ScreenID, seatIDs)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ValidationError{"one or more seats do not belong to this showtime's screen"}
	}

	booked, err := s.booked.BookedAmong(ctx, showtimeID, seatIDs)
	if err != nil {
		return nil, err
	}
	if len(booked) > 0 {
		return nil, SeatsUnavailableError{SeatIDs: booked}
	}

	conflicts, err := s.holder.Hold(ctx, showtimeID, seatIDs, userID)
	if err != nil {
		return nil, err
	}
	if len(conflicts) > 0 {
		return nil, SeatsUnavailableError{SeatIDs: conflicts}
	}

	return &Result{
		ShowtimeID: showtimeID,
		SeatIDs:    seatIDs,
		ExpiresAt:  s.now().Add(s.ttl),
	}, nil
}

// cleanSeatIDs: 1..6 ids, lowercase, trimmed, no duplicates.
func cleanSeatIDs(in []string) ([]string, error) {
	if len(in) == 0 {
		return nil, ValidationError{"seat_ids is required"}
	}
	if len(in) > MaxSeatsPerHold {
		return nil, ValidationError{fmt.Sprintf("you can hold at most %d seats at once", MaxSeatsPerHold)}
	}

	seen := make(map[string]bool, len(in))
	out := make([]string, 0, len(in))
	for _, id := range in {
		id = strings.ToLower(strings.TrimSpace(id))
		if id == "" {
			return nil, ValidationError{"seat_ids must not contain empty values"}
		}
		if seen[id] {
			return nil, ValidationError{"seat_ids must not contain duplicates"}
		}
		seen[id] = true
		out = append(out, id)
	}
	return out, nil
}
