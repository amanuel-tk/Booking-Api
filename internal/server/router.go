package server

import (
	"context"
	"net/http"
	"time"

	"github.com/amanuel-tk/Booking-Api/internal/auth"
	"github.com/amanuel-tk/Booking-Api/internal/httpx"
	"github.com/amanuel-tk/Booking-Api/internal/movie"
	"github.com/amanuel-tk/Booking-Api/internal/showtime"
	"github.com/amanuel-tk/Booking-Api/internal/theater"
	"github.com/amanuel-tk/Booking-Api/internal/user"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Deps struct {
	DB       *pgxpool.Pool
	RDB      *redis.Client
	JWT      *auth.JWTManager
	User     *user.Handler
	Movie    *movie.Handler
	Theater  *theater.Handler
	Showtime *showtime.Handler
}

func NewRouter(d Deps) http.Handler {
	r := chi.NewRouter()

	r.Use(
		middleware.RequestID,
		middleware.RealIP,
		middleware.Logger,
		middleware.Recoverer,
		middleware.Timeout(30*time.Second),
	)
	r.Get("/health", healthHandler(d))

	r.Route("/api/v1", func(r chi.Router) {

		r.Post("/auth/register", d.User.Register)
		r.Post("/auth/login", d.User.Login)

		r.Get("/movies", d.Movie.ListMovies)
		r.Get("/movies/{id}", d.Movie.GetMovie)

		r.Get("/theaters", d.Theater.GetTheaters)
		r.Get("/theaters/{id}/screens", d.Theater.ListScreens)
		r.Get("/screens/{id}", d.Theater.GetScreen)

		r.Group(func(r chi.Router) {
			r.Use(d.JWT.Authenticate)

			r.Get("/me", d.User.Me)

			r.Group(func(r chi.Router) {
				r.Use(auth.RequireRole("admin"))

				r.Post("/movies", d.Movie.Create)
				r.Post("/theaters", d.Theater.CreateTheater)
				r.Post("/theaters/{id}/screens", d.Theater.CreateScreen)
				r.Post("/showtimes", d.Showtime.CreateShowtime)

			})
		})
	})

	return r

}

func healthHandler(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}
