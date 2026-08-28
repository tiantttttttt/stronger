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

	"stronger/server/internal/auth"
	"stronger/server/internal/platform/config"
	"stronger/server/internal/platform/database"
	platformhttp "stronger/server/internal/platform/http"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx := context.Background()
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	tokens := auth.NewTokenManager(cfg.JWTSecret, time.Hour)
	devUsers := database.NewDevUsers(pool)
	workouts := database.NewWorkouts(pool)

	server := &http.Server{
		Addr: cfg.HTTPAddr,
		Handler: platformhttp.NewRouter(platformhttp.Dependencies{
			Database:    pool,
			Environment: cfg.Environment,
			Tokens:      tokens,
			DevUsers:    devUsers,
			Workouts:    workouts,
		}),
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverError := make(chan error, 1)
	go func() {
		log.Printf("stronger-api listening on %s (%s)", cfg.HTTPAddr, cfg.Environment)
		serverError <- server.ListenAndServe()
	}()

	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(shutdownSignal)

	select {
	case <-shutdownSignal:
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-serverError:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}
