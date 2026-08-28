package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DevUsers prepares deterministic local-only users for the development login route.
type DevUsers struct {
	pool *pgxpool.Pool
}

func NewDevUsers(pool *pgxpool.Pool) *DevUsers {
	return &DevUsers{pool: pool}
}

func (users *DevUsers) EnsureDevUser(ctx context.Context, userID string) error {
	_, err := users.pool.Exec(ctx, `
		INSERT INTO users (
			id, wechat_openid, nickname, height_cm, weight_kg,
			training_level, fitness_goal, timezone, last_login_at
		) VALUES ($1, $2, 'Development User', 170, 65, 1, 3, 'Asia/Shanghai', now())
		ON CONFLICT (id) DO UPDATE SET
			last_login_at = now(),
			updated_at = now()
	`, userID, "dev:"+userID)
	return err
}
