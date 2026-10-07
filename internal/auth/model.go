package auth

import "time"

type RefreshToken struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	TokenHash string     `json:"-"`
	FamilyID  string     `json:"family_id"`
	ExpiresAt time.Time  `json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at"`
	UserAgent string     `json:"user_agent"`
	IP        string     `json:"ip"`
	CreatedAt time.Time  `json:"created_at"`
}
