package theater

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

type CreateTheater struct {
	Name string `json:"name"`
	City string `json:"city"`
}
type createScreenRequest struct {
	Name        string `json:"name"`
	Rows        int    `json:"rows"`
	SeatsPerRow int    `json:"seats_per_row"`
}

const (
	MaxRows        = 26
	MaxSeatsPerRow = 50
	maxNameLen     = 100
)

func (s *Service) CreateTheater(ctx context.Context, req CreateTheater) (*Theater, error) {

	name := strings.TrimSpace(strings.ToLower(req.Name))
	city := strings.TrimSpace(strings.ToLower(req.City))

	if name == "" || city == "" {
		return nil, ValidationError{"name and city are required"}
	}

	if len(name) > maxNameLen || len(city) > maxNameLen {
		return nil, ValidationError{"name and city must be at most 100 characters"}
	}

	t := &Theater{Name: name, City: city}

	if err := s.repo.CreateTheater(ctx, t); err != nil {
		return nil, err
	}
	return t, nil

}

func (s *Service) CreateScreen(ctx context.Context, theaterID string, req createScreenRequest) (*Screen, error) {

	name := strings.TrimSpace(strings.ToLower(req.Name))

	if name == "" {
		return nil, ValidationError{"screen name is required"}
	}
	if len(name) > maxNameLen {
		return nil, ValidationError{"screen name must be at most 100 characters"}
	}
	if _, err := uuid.Parse(theaterID); err != nil {
		return nil, ErrNotFound
	}

	seats, err := GenerateSeats(req.Rows, req.SeatsPerRow)

	if err != nil {
		return nil, err
	}

	screen := &Screen{
		TheaterID:   theaterID,
		Name:        name,
		RowCount:    req.Rows,
		SeatsPerRow: req.SeatsPerRow,
	}

	if err := s.repo.CreateScreenWithSeats(ctx, screen, seats); err != nil {
		return nil, err
	}
	return screen, nil
}

func (s *Service) GetTheaters(ctx context.Context) ([]Theater, error) {
	return s.repo.GetTheaters(ctx)
}

func (s *Service) GetScreen(ctx context.Context, id string) (*ScreenLayout, error) {
	screen, err := s.repo.GetScreen(ctx, id)
	if err != nil {
		return nil, err
	}

	seats, err := s.repo.ListSeats(ctx, id)
	if err != nil {
		return nil, err
	}

	layout := &ScreenLayout{Screen: *screen, Rows: []SeatRow{}}
	for _, seat := range seats {
		n := len(layout.Rows)
		if n == 0 || layout.Rows[n-1].Row != seat.RowLabel {
			layout.Rows = append(layout.Rows, SeatRow{Row: seat.RowLabel})
			n++
		}
		layout.Rows[n-1].Seats = append(layout.Rows[n-1].Seats, seat)
	}
	return layout, nil
}

func GenerateSeats(rows, seatsPerRow int) ([]Seat, error) {
	if rows < 1 || rows > MaxRows {
		return nil, ValidationError{"rows must be between 1 and 26"}
	}
	if seatsPerRow < 1 || seatsPerRow > MaxSeatsPerRow {
		return nil, ValidationError{"seats_per_row must be between 1 and 50"}
	}

	seats := make([]Seat, 0, rows*seatsPerRow)
	for r := 0; r < rows; r++ {
		label := string(rune('A' + r))

		for n := 1; n < seatsPerRow; n++ {
			seats = append(seats, Seat{RowLabel: label, SeatNumber: n})
		}
	}
	return seats, nil
}
