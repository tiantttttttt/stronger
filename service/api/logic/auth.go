package logic

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"stronger/server/consts"
	"stronger/server/service/model"
)

type Auth struct{}

var AuthLogic = &Auth{}

// EnsureDevUser 为开发登录准备可重复使用的测试用户。
func (a *Auth) EnsureDevUser(ctx context.Context, userID string) error {
	height, weight := "170", "65"
	level, goal := int16(1), int16(3)
	user := model.User{
		ID: userID, WechatOpenID: "dev:" + userID, Nickname: "Development User",
		HeightCM: &height, WeightKG: &weight, TrainingLevel: &level, FitnessGoal: &goal,
		Timezone: "Asia/Shanghai",
	}
	// 插入及冲突更新均使用数据库时间，保持时间约束一致。
	return consts.DB.WithContext(ctx).Model(&model.User{}).Clauses(clause.OnConflict{
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
