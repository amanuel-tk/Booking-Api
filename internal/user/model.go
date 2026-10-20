package user

import "time"

type LoginResponse struct {
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIs..."`
	TokenType   string `json:"token_type" example:"Bearer"`
	User        User   `json:"user"`
}

type User struct {
	ID           string    `json:"id" example:"3f2b8c1e-5a7d-4e6f-9b1a-2c3d4e5f6a7b" format:"uuid"`
	FirstName    string    `json:"first_name" example:"jane"`
	LastName     string    `json:"last_name" example:"doe"`
	Email        string    `json:"email" example:"jane@example.com"`
	Roles        []string  `json:"roles" example:"user"`
	Permissions  []string  `json:"permissions"`
	IsActive     bool      `json:"is_active"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
