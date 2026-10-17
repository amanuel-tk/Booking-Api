package booking

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/amanuel-tk/Booking-Api/internal/showtime"
)

const (
	CancelCutoff       = 2 * time.Hour
	MaxSeatsPerBooking = 6
)

type ShowtimeGetter interface {
	GetShowtime(ctx context.Context, id string) (*showtime.Showtime, error)
}

type SeatChecker interface {
	SeatsBelongToScreen(ctx context.Context, screenID string, seatIDs []string) (bool, error)
}

type HoldStore interface {
	Owned(ctx context.Context, showtimeID string, seatIDs []string, userID string) (bool, error)
	Release(ctx context.Context, showtimeID string, seatIDs []string, userID string) error
}

type Repo interface {
	CreateBooking(ctx context.Context, nb *NewBooking) error
	GetForUser(ctx context.Context, bookingID, userID string) (*BookingView, error)
	ListForUser(ctx context.Context, userID string) ([]BookingView, error)
	CancelBooking(ctx context.Context, bookingID, userID string) error
}

type Service struct {
	repo      Repo
	showtimes ShowtimeGetter
	seats     SeatChecker
	holds     HoldStore
	now       func() time.Time
}

func NewService(repo Repo, showtimes ShowtimeGetter, seats SeatChecker, holds HoldStore) *Service {
	return &Service{repo: repo, showtimes: showtimes, seats: seats, holds: holds, now: time.Now}
}

func (s *Service) Create(ctx context.Context, userID, showtimeID string, seatIDs []string) (*BookingView, error) {
	seatIDs, err := cleanSeatIDs(seatIDs)
	if err != nil {
		return nil, err
	}

	st, err := s.showtimes.GetShowtime(ctx, showtimeID)
	if err != nil {
		if errors.Is(err, showtime.ErrNotFound) {
			return nil, ErrShowtimeNotFound
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

	owned, err := s.holds.Owned(ctx, showtimeID, seatIDs, userID)
	if err != nil {
		return nil, err
	}
	if !owned {
		return nil, ErrHoldNotOwned
	}

	nb := &NewBooking{
		UserID:     userID,
		ShowtimeID: showtimeID,
		SeatIDs:    seatIDs,
		PriceCents: st.PriceCents,
		TotalCents: st.PriceCents * len(seatIDs),
	}
	if err := s.repo.CreateBooking(ctx, nb); err != nil {
		return nil, err
	}

	if err := s.holds.Release(ctx, showtimeID, seatIDs, userID); err != nil {
		log.Printf("booking %s: release holds: %v", nb.ID, err)
	}

	return s.repo.GetForUser(ctx, nb.ID, userID)
}

func (s *Service) ListMine(ctx context.Context, userID string) ([]BookingView, error) {
	return s.repo.ListForUser(ctx, userID)
}

func (s *Service) GetMine(ctx context.Context, userID, bookingID string) (*BookingView, error) {
	return s.repo.GetForUser(ctx, bookingID, userID)
}

func (s *Service) Cancel(ctx context.Context, userID, bookingID string) (*BookingView, error) {
	v, err := s.repo.GetForUser(ctx, bookingID, userID)
	if err != nil {
		return nil, err
	}
	if v.Status == StatusCancelled {
		return nil, ErrAlreadyCancelled
	}
	if !CanCancel(v.StartsAt, s.now()) {
		return nil, ErrTooLateToCancel
	}

	if err := s.repo.CancelBooking(ctx, bookingID, userID); err != nil {
		return nil, err
	}
	return s.repo.GetForUser(ctx, bookingID, userID)
}

func CanCancel(startsAt, now time.Time) bool {
	return startsAt.Sub(now) >= CancelCutoff
}

func cleanSeatIDs(in []string) ([]string, error) {
	if len(in) == 0 {
		return nil, ValidationError{"seat_ids is required"}
	}
	if len(in) > MaxSeatsPerBooking {
		return nil, ValidationError{fmt.Sprintf("you can book at most %d seats at once", MaxSeatsPerBooking)}
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
