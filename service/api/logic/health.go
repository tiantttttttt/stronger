package logic

import (
	"context"
	"errors"

	"stronger/server/consts"
)

type Health struct{}

var HealthLogic = &Health{}

func (h *Health) Ping(ctx context.Context) error {
	if consts.DB == nil {
		return errors.New("database is not initialized")
	}
	pool, err := consts.DB.DB()
	if err != nil {
		return err
	}
	return pool.PingContext(ctx)
}
