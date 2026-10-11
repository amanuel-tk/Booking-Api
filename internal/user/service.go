package user

import (
	"context"
	"net/mail"
	"strings"

	"github.com/amanuel-tk/Booking-Api/internal/auth"
)

type RegisterRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type ValidationError struct {
	Msg string
}

func (e *ValidationError) Error() string {
	return e.Msg
}

type Service struct {
	users *Repository
	jwt   *auth.JWTManager
}

func NewService(users *Repository, jwt *auth.JWTManager) *Service {

	return &Service{
		users: users,
		jwt:   jwt,
	}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*User, error) {
	first := strings.ToLower(strings.TrimSpace(req.FirstName))
	last := strings.ToLower(strings.TrimSpace(req.LastName))
	email := strings.ToLower(strings.TrimSpace(req.Email))

	if first == "" || last == "" {
		return nil, &ValidationError{"first name and last name are required"}
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, &ValidationError{"invalid email"}
	}
	// bcrypt only uses the first 72 bytes, so cap it
	if len(req.Password) < 8 || len(req.Password) > 72 {
		return nil, &ValidationError{"password must be 8 to 72 characters"}
	}

	hash, err := auth.HashPassword(req.Password)

	if err != nil {
		return nil, err
	}

	u := &User{
		FirstName:    first,
		LastName:     last,
		Email:        email,
		PasswordHash: hash,
	}
	if err := s.users.Create(ctx, u, "user"); err != nil {
		return nil, err
	}

	u.Roles = []string{"user"}
	return u, nil
}
