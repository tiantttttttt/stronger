package responses

import "time"

type WorkoutSummary struct {
	ID                     string     `json:"id"`
	Status                 int16      `json:"status"`
	StartedAt              time.Time  `json:"started_at"`
	CompletedAt            *time.Time `json:"completed_at,omitempty"`
	DurationSeconds        *int       `json:"duration_seconds,omitempty"`
	CompletedExerciseCount *int       `json:"completed_exercise_count,omitempty"`
	CompletedSetCount      *int       `json:"completed_set_count,omitempty"`
	TotalReps              *int64     `json:"total_reps,omitempty"`
	TotalVolumeKG          *string    `json:"total_volume_kg,omitempty"`
}
