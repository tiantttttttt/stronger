package responses

import "time"

// WorkoutRecord 为写入响应，不装配统计数据。
type WorkoutRecord struct {
	ID          string     `json:"id"`
	Status      int16      `json:"status"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type WorkoutStat struct {
	WorkoutRecord
	DurationSeconds        *int    `json:"duration_seconds,omitempty"`
	CompletedExerciseCount *int    `json:"completed_exercise_count,omitempty"`
	CompletedSetCount      *int    `json:"completed_set_count,omitempty"`
	TotalReps              *int64  `json:"total_reps,omitempty"`
	TotalVolumeKG          *string `json:"total_volume_kg,omitempty"`
}
