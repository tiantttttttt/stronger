package model

import "time"

type WorkoutStatus int16

const (
	WorkoutActive    WorkoutStatus = 1
	WorkoutCompleted WorkoutStatus = 2
	WorkoutDiscarded WorkoutStatus = 3
)

// WorkoutSession maps the currently used columns; it deliberately has no soft delete.
type WorkoutSession struct {
	ID                     string        `gorm:"type:uuid;primaryKey"`
	UserID                 string        `gorm:"type:uuid"`
	Status                 WorkoutStatus `gorm:"default:1"`
	StartedAt              time.Time     `gorm:"default:now()"`
	CompletedAt            *time.Time
	DurationSeconds        *int
	CompletedExerciseCount *int
	CompletedSetCount      *int
	TotalReps              *int64
	TotalVolumeKG          *string   `gorm:"column:total_volume_kg;type:numeric(16,2)"`
	CreatedAt              time.Time `gorm:"autoCreateTime:false;default:now()"`
	UpdatedAt              time.Time `gorm:"autoUpdateTime:false;default:now()"`
}

func (WorkoutSession) TableName() string { return "workout_sessions" }
