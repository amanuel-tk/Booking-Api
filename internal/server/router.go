package server

import (
	"context"
	"net/http"
	"time"

	"github.com/amanuel-tk/Booking-Api/internal/auth"
	"github.com/amanuel-tk/Booking-Api/internal/httpx"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Deps struct {
	DB  *pgxpool.Pool
	RDB *redis.Client
	JWT *auth.JWTManager
}

func NewRouter(d Deps) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID, middleware.RealIP, middleware.Logger, middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)

		defer cancel()

		status := map[string]string{"postgres": "ok", "redis": "ok"}
		code := http.StatusOK

		if err := d.DB.Ping(ctx); err != nil {
			status["postgres"], code = "down", http.StatusServiceUnavailable
		}

		if err := d.RDB.Ping(ctx).Err(); err != nil {
			status["redis"], code = "down", http.StatusServiceUnavailable
		}

		httpx.WriteJSON(w, code, status)

	})

	return r

}
