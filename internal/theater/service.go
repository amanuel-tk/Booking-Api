package theater

import (
	"context"
	"strings"
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
