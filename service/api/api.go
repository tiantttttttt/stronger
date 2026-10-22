package api

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"time"

	"stronger/server/consts"
	"stronger/server/pkg/db"
	"stronger/server/service/api/middleware"
	"stronger/server/service/api/router"
	"stronger/server/service/config"
)

// Service 负责 HTTP 服务与数据库连接池的生命周期。
type Service struct {
	server *http.Server
	pool   *sql.DB
	debug  bool
}

func New(cfg config.Config) (*Service, error) {
	database, err := db.NewDB(context.Background(), cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	pool, err := database.DB()
	if err != nil {
		return nil, err
	}
	consts.DB = database
	tokens := middleware.NewTokenManager(cfg.JWTSecret, time.Hour)
	return &Service{
		pool: pool, debug: cfg.Debug,
		server: &http.Server{
			Addr: cfg.HTTPAddr, ReadHeaderTimeout: 5 * time.Second,
			Handler: router.NewRouter(cfg.Debug, tokens),
		},
	}, nil
}

// Run 在取消或监听失败后停止服务并释放资源。
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

// Close 先关闭 HTTP 服务，再释放数据库连接池。
func (s *Service) Close(ctx context.Context) error {
	shutdownErr := s.server.Shutdown(ctx)
	if shutdownErr != nil {
		shutdownErr = errors.Join(shutdownErr, s.server.Close())
	}
	return errors.Join(shutdownErr, s.pool.Close())
}
