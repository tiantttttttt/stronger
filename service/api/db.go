package api

import (
	"context"
	"database/sql"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewDB opens GORM and its underlying connection pool. /health checks connectivity.
// Schema changes are applied through SQL migrations, never AutoMigrate.
func NewDB(ctx context.Context, databaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		DisableAutomaticPing: true,
		Logger:               logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}
	pool, err := db.DB()
	if err != nil {
		return nil, err
	}
	pool.SetMaxOpenConns(10)
	pool.SetMaxIdleConns(5)
	pool.SetConnMaxLifetime(time.Hour)
	return db.WithContext(ctx), nil
}

type databasePinger struct{ pool *sql.DB }

func (p databasePinger) Ping(ctx context.Context) error { return p.pool.PingContext(ctx) }
