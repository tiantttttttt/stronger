package logic

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"stronger/server/consts"
	"stronger/server/service/api/logic/responses"
	"stronger/server/service/model"
)

type Workouts struct{}

var WorkoutsLogic = &Workouts{}

func (w *Workouts) Start(ctx context.Context, userID string) (responses.WorkoutSummary, error) {
	var result model.WorkoutSession
	err := consts.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user model.User
		if err := tx.Select("id", "height_cm", "weight_kg", "training_level", "fitness_goal").
			Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", userID).Take(&user).Error; err != nil {
			return workoutError(err)
		}
		if !user.ProfileComplete() {
			return model.ErrProfileIncomplete
		}
		err := tx.Select("id", "status", "started_at").
			Where("user_id = ? AND status = ?", userID, model.WorkoutActive).Take(&result).Error
		if err == nil {
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		result = model.WorkoutSession{ID: id.String(), UserID: userID}
		// 由 PostgreSQL 填充时间和状态，并回读默认值。
		return tx.Select("id", "user_id").Clauses(clause.Returning{}).Create(&result).Error
	})
	return workoutSummary(result), err
}

func (w *Workouts) Complete(ctx context.Context, userID, workoutID string) (responses.WorkoutSummary, error) {
	var result model.WorkoutSession
	err := consts.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND user_id = ?", workoutID, userID).Take(&result).Error
		if err != nil {
			return workoutError(err)
		}
		if result.Status == model.WorkoutCompleted {
			return nil
		}
		if result.Status != model.WorkoutActive {
			return model.ErrTerminalState
		}
		// 当前仅支持空训练，实际组次汇总随组次功能实现。
		return tx.Model(&result).Clauses(clause.Returning{}).
			Where("id = ? AND user_id = ? AND status = ?", workoutID, userID, model.WorkoutActive).
			Updates(map[string]interface{}{
				"status":                   model.WorkoutCompleted,
				"completed_at":             gorm.Expr("now()"),
				"duration_seconds":         gorm.Expr("GREATEST(0, EXTRACT(EPOCH FROM (now() - started_at))::integer)"),
				"completed_exercise_count": 0, "completed_set_count": 0,
				"total_reps": 0, "total_volume_kg": 0, "updated_at": gorm.Expr("now()"),
			}).Error
	})
	return workoutSummary(result), err
}

func (w *Workouts) GetSummary(ctx context.Context, userID, workoutID string) (responses.WorkoutSummary, error) {
	var result model.WorkoutSession
	err := consts.DB.WithContext(ctx).
		Where("id = ? AND user_id = ? AND status = ?", workoutID, userID, model.WorkoutCompleted).
		Take(&result).Error
	return workoutSummary(result), workoutError(err)
}

func workoutError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.ErrNotFound
	}
	return err
}

func workoutSummary(s model.WorkoutSession) responses.WorkoutSummary {
	return responses.WorkoutSummary{
		ID: s.ID, Status: int16(s.Status), StartedAt: s.StartedAt, CompletedAt: s.CompletedAt,
		DurationSeconds: s.DurationSeconds, CompletedExerciseCount: s.CompletedExerciseCount,
		CompletedSetCount: s.CompletedSetCount, TotalReps: s.TotalReps, TotalVolumeKG: s.TotalVolumeKG,
	}
}
