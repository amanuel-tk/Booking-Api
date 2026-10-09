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
}

func LoadConfig() (*Config, error) {

	_ = godotenv.Load()

	ttl, err := time.ParseDuration(getEnv("ACCESS_TOKEN_TTL", "15m"))

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
	}

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
