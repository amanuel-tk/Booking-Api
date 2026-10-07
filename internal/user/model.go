package user

import "time"

type User struct {
	ID           string    `json:"id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Email        string    `json:"email"`
	Roles        []string  `json:"roles"`
	Permissions  []string  `json:"permissions"`
	IsActive     bool      `json:"is_active"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
