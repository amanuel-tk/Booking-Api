package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	DatabaseURL    string
	RedisAddr      string
	RedisPassword  string
	JWTSecret      string
	AccessTokenTTL time.Duration
	CinemaTZ       string
	CinemaLoc      *time.Location
	HoldTTL        time.Duration
}

func LoadConfig() (*Config, error) {

	_ = godotenv.Load()

	ttl, err := time.ParseDuration(getEnv("ACCESS_TOKEN_TTL", "15m"))
	holdTTL, err := time.ParseDuration(getEnv("HOLD_TTL", "10m"))
	if err != nil {
		return nil, fmt.Errorf("invalid HOLD_TTL: %w", err)
	}
	if holdTTL <= 0 {
		return nil, fmt.Errorf("HOLD_TTL must be positive")
	}

	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Port:           getEnv("PORT", "8080"),
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		RedisAddr:      getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:  getEnv("REDIS_PASSWORD", ""),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		AccessTokenTTL: ttl,
		HoldTTL:        holdTTL,
		CinemaTZ:       getEnv("CINEMA_TZ", "Africa/Addis_Ababa"),
	}

	loc, err := time.LoadLocation(cfg.CinemaTZ)
	if err != nil {
		return nil, fmt.Errorf("load time zone: %v", err)
	}
	cfg.CinemaLoc = loc

	if cfg.DatabaseURL == "" || cfg.JWTSecret == "" {
		return nil, fmt.Errorf("DATABASE_URL and JWT_SECRET are required")
	}

	return cfg, nil
}

func getEnv(key string, defaultValue string) string {

	if v := os.Getenv(key); v != "" {
		return v
	}

	return defaultValue

}
