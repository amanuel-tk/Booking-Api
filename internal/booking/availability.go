package booking

import (
	"context"
	"errors"

	"github.com/amanuel-tk/Booking-Api/internal/showtime"
	"github.com/amanuel-tk/Booking-Api/internal/theater"
)

type SeatLister interface {
	ListSeats(ctx context.Context, screenID string) ([]theater.Seat, error)
}

type BookedReader interface {
	BookedSeatIDs(ctx context.Context, showtimeID string) (map[string]bool, error)
}

type HeldReader interface {
	Held(ctx context.Context, showtimeID string, seatIDs []string) (map[string]bool, error)
}

type AvailabilityService struct {
	showtimes ShowtimeGetter
	seats     SeatLister
	booked    BookedReader
	held      HeldReader
}

func NewAvailabilityService(showtimes ShowtimeGetter, seats SeatLister, booked BookedReader, held HeldReader) *AvailabilityService {
	return &AvailabilityService{showtimes: showtimes, seats: seats, booked: booked, held: held}
}

func (a *AvailabilityService) SeatMap(ctx context.Context, showtimeID string) (*SeatMap, error) {
	st, err := a.showtimes.GetShowtime(ctx, showtimeID)
	if err != nil {
		if errors.Is(err, showtime.ErrNotFound) {
			return nil, ErrShowtimeNotFound
		}
		return nil, err
	}

	seats, err := a.seats.ListSeats(ctx, st.ScreenID)
	if err != nil {
		return nil, err
	}

	ids := make([]string, len(seats))
	for i, s := range seats {
		ids[i] = s.ID
	}

	booked, err := a.booked.BookedSeatIDs(ctx, showtimeID)
	if err != nil {
		return nil, err
	}
	held, err := a.held.Held(ctx, showtimeID, ids)
	if err != nil {
		return nil, err
	}

	return buildSeatMap(showtimeID, st.ScreenID, seats, booked, held), nil
}

func buildSeatMap(showtimeID, screenID string, seats []theater.Seat, booked, held map[string]bool) *SeatMap {
	m := &SeatMap{ShowtimeID: showtimeID, ScreenID: screenID, Rows: []SeatMapRow{}}

	for _, s := range seats {
		status := SeatFree
		switch {
		case booked[s.ID]:
			status = SeatBooked
		case held[s.ID]:
			status = SeatHeld
		}
		if status == SeatFree {
			m.Available++
		}

		n := len(m.Rows)
		if n == 0 || m.Rows[n-1].Row != s.RowLabel {
			m.Rows = append(m.Rows, SeatMapRow{Row: s.RowLabel})
			n++
		}
		m.Rows[n-1].Seats = append(m.Rows[n-1].Seats, SeatView{
			ID:     s.ID,
			Number: s.SeatNumber,
			Status: status,
		})
	}
	return m
}
