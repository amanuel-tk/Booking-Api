package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Roles       []string `json:"roles"`
	Permissions []string `json:"perm,omitempty"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secret []byte
	ttl    time.Duration
}

func NewJWTManager(secret string, ttl time.Duration) *JWTManager {
	return &JWTManager{
		secret: []byte(secret),
		ttl:    ttl}
}

func (m *JWTManager) Generate(userID string, roles, perms []string) (string, error) {
	now := time.Now()
	claims := Claims{
		Roles:       roles,
		Permissions: perms,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodES256, claims).SignedString(m.secret)
}

func (m *JWTManager) Parse(raw string) (*Claims, error) {

	claims := &Claims{}

	token, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
		return m.secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil

}
