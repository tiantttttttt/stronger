package api

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"time"

	"stronger/server/service/api/logic"
	"stronger/server/service/api/middleware"
	"stronger/server/service/api/router"
	"stronger/server/service/config"
)

// Service owns HTTP and database resources for the API.
type Service struct {
	server *http.Server
	pool   *sql.DB
	debug  bool
}

func New(cfg config.Config) (*Service, error) {
	db, err := NewDB(context.Background(), cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	pool, err := db.DB()
	if err != nil {
		return nil, err
	}
	tokens := middleware.NewTokenManager(cfg.JWTSecret, time.Hour)
	return &Service{
		pool: pool, debug: cfg.Debug,
		server: &http.Server{
			Addr: cfg.HTTPAddr, ReadHeaderTimeout: 5 * time.Second,
			Handler: router.NewRouter(router.Dependencies{
				Database: databasePinger{pool: pool}, Debug: cfg.Debug,
				Tokens: tokens, DevUsers: logic.NewAuth(db), Workouts: logic.NewWorkouts(db),
			}),
		},
	}, nil
}

// Run serves requests until cancellation or a listener error, then releases resources.
func (s *Service) Run(ctx context.Context) error {
	serverError := make(chan error, 1)
	go func() {
		log.Printf("stronger-api listening on %s (debug=%t)", s.server.Addr, s.debug)
		serverError <- s.server.ListenAndServe()
	}()
	var runErr error
	listenerStopped := false
	select {
	case <-ctx.Done():
	case runErr = <-serverError:
		listenerStopped = true
		if errors.Is(runErr, http.ErrServerClosed) {
			runErr = nil
		}
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	closeErr := s.Close(shutdownCtx)
	if !listenerStopped {
		runErr = <-serverError
		if errors.Is(runErr, http.ErrServerClosed) {
			runErr = nil
		}
	}
	return errors.Join(runErr, closeErr)
}

// Close drains HTTP requests before closing the database pool.
func (s *Service) Close(ctx context.Context) error {
	shutdownErr := s.server.Shutdown(ctx)
	if shutdownErr != nil {
		shutdownErr = errors.Join(shutdownErr, s.server.Close())
	}
	return errors.Join(shutdownErr, s.pool.Close())
}
