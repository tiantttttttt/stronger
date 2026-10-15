package logic

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"stronger/server/service/model"
)

type Auth struct{ db *gorm.DB }

func NewAuth(db *gorm.DB) *Auth { return &Auth{db: db} }

// EnsureDevUser prepares deterministic local-only users for development login.
func (a *Auth) EnsureDevUser(ctx context.Context, userID string) error {
	height, weight := "170", "65"
	level, goal := int16(1), int16(3)
	user := model.User{
		ID: userID, WechatOpenID: "dev:" + userID, Nickname: "Development User",
		HeightCM: &height, WeightKG: &weight, TrainingLevel: &level, FitnessGoal: &goal,
		Timezone: "Asia/Shanghai",
	}
	// Use database timestamps, including on conflicts, to preserve the migration's checks.
	return a.db.WithContext(ctx).Model(&model.User{}).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"last_login_at": gorm.Expr("now()"), "updated_at": gorm.Expr("now()"),
		}),
	}).Create(map[string]interface{}{
		"id": user.ID, "wechat_openid": user.WechatOpenID, "nickname": user.Nickname,
		"height_cm": height, "weight_kg": weight,
		"training_level": level, "fitness_goal": goal, "timezone": user.Timezone,
		"last_login_at": gorm.Expr("now()"),
	}).Error
}
