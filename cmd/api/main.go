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

	"github.com/amanuel-tk/Booking-Api/internal/auth"
	"github.com/amanuel-tk/Booking-Api/internal/config"
	"github.com/amanuel-tk/Booking-Api/internal/database"
	"github.com/amanuel-tk/Booking-Api/internal/server"
)

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

	srv := &http.Server{
		Addr: ":" + cfg.Port,
		Handler: server.NewRouter(server.Deps{
			DB:  db,
			RDB: rdb,
			JWT: jwtm,
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
