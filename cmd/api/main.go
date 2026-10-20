package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata"

	"github.com/amanuel-tk/Booking-Api/internal/auth"
	"github.com/amanuel-tk/Booking-Api/internal/booking"
	"github.com/amanuel-tk/Booking-Api/internal/config"
	"github.com/amanuel-tk/Booking-Api/internal/database"
	"github.com/amanuel-tk/Booking-Api/internal/hold"
	"github.com/amanuel-tk/Booking-Api/internal/movie"
	"github.com/amanuel-tk/Booking-Api/internal/server"
	"github.com/amanuel-tk/Booking-Api/internal/showtime"
	"github.com/amanuel-tk/Booking-Api/internal/theater"
	"github.com/amanuel-tk/Booking-Api/internal/user"

	_ "github.com/amanuel-tk/Booking-Api/docs"
)

// @title           Booking API
// @version         1.0
// @description     Movie booking API: browse movies and showtimes, hold seats, book tickets.
// @BasePath        /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer " followed by your JWT.

func main() {

	cfg, err := config.LoadConfig()

	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := database.NewPostgres(ctx, cfg.DatabaseURL)

	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rdb, err := database.NewRedis(ctx, cfg.RedisAddr, cfg.RedisPassword)

	if err != nil {
		log.Fatal(err)
	}

	defer rdb.Close()

	jwtm := auth.NewJWTManager(cfg.JWTSecret, cfg.AccessTokenTTL)

	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo, jwtm)
	userHandler := user.NewHandler(userService)

	movieRepo := movie.NewRepository(db)
	movieService := movie.NewService(movieRepo)
	movieHandler := movie.NewHandler(movieService)

	theaterRepo := theater.NewRepository(db)
	theaterService := theater.NewService(theaterRepo)
	theaterHandler := theater.NewHandler(theaterService)

	showtimeRepo := showtime.NewRepository(db)
	showtimeService := showtime.NewService(showtimeRepo, movieService, theaterService, cfg.CinemaLoc)
	showtimeHandler := showtime.NewHandler(showtimeService)

	holdStore := hold.NewStore(rdb, cfg.HoldTTL)
	bookingRepo := booking.NewRepository(db)

	holdSvc := hold.NewService(holdStore, showtimeService, theaterService, bookingRepo, cfg.HoldTTL)
	holdHandler := hold.NewHandler(holdSvc)

	bookingSvc := booking.NewService(bookingRepo, showtimeService, theaterService, holdStore)
	availSvc := booking.NewAvailabilityService(showtimeService, theaterService, bookingRepo, holdStore)
	bookingHandler := booking.NewHandler(bookingSvc, availSvc)

	srv := &http.Server{
		Addr: ":" + cfg.Port,
		Handler: server.NewRouter(server.Deps{
			DB:             db,
			RDB:            rdb,
			JWT:            jwtm,
			User:           userHandler,
			Movie:          movieHandler,
			Theater:        theaterHandler,
			Showtime:       showtimeHandler,
			Hold:           holdHandler,
			Booking:        bookingHandler,
			SwaggerEnabled: true,
		}),
		ReadTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_ = srv.Shutdown(shutdownCtx)

}
