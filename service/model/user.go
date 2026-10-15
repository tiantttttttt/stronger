package model

import "time"

// User maps the columns needed by the current development login and profile gate.
// SQL migrations remain responsible for schema defaults and constraints.
type User struct {
	ID            string `gorm:"type:uuid;primaryKey"`
	WechatOpenID  string `gorm:"column:wechat_openid"`
	Nickname      string
	HeightCM      *string `gorm:"column:height_cm"`
	WeightKG      *string `gorm:"column:weight_kg"`
	TrainingLevel *int16
	FitnessGoal   *int16
	Timezone      string
	LastLoginAt   time.Time
	CreatedAt     time.Time `gorm:"autoCreateTime:false;default:now()"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime:false;default:now()"`
}

func (User) TableName() string { return "users" }

func (u User) ProfileComplete() bool {
	return u.HeightCM != nil && u.WeightKG != nil && u.TrainingLevel != nil && u.FitnessGoal != nil
}
