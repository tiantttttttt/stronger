package database

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"stronger/server/internal/workout"
)

type Workouts struct {
	pool *pgxpool.Pool
}

func NewWorkouts(pool *pgxpool.Pool) *Workouts {
	return &Workouts{pool: pool}
}

func (repository *Workouts) Start(ctx context.Context, userID string) (workout.Summary, error) {
	tx, err := repository.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return workout.Summary{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var profileComplete bool
	err = tx.QueryRow(ctx, `
		SELECT height_cm IS NOT NULL
			AND weight_kg IS NOT NULL
			AND training_level IS NOT NULL
			AND fitness_goal IS NOT NULL
		FROM users
		WHERE id = $1
		FOR UPDATE
	`, userID).Scan(&profileComplete)
	if errors.Is(err, pgx.ErrNoRows) {
		return workout.Summary{}, workout.ErrNotFound
	}
	if err != nil {
		return workout.Summary{}, err
	}
	if !profileComplete {
		return workout.Summary{}, workout.ErrProfileIncomplete
	}

	result, err := getActiveWorkout(ctx, tx, userID)
	if err == nil {
		if err := tx.Commit(ctx); err != nil {
			return workout.Summary{}, err
		}
		return result, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return workout.Summary{}, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return workout.Summary{}, err
	}
	result = workout.Summary{ID: id.String()}
	err = tx.QueryRow(ctx, `
		INSERT INTO workout_sessions (id, user_id)
		VALUES ($1, $2)
		RETURNING status, started_at
	`, result.ID, userID).Scan(&result.Status, &result.StartedAt)
	if err != nil {
		return workout.Summary{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return workout.Summary{}, err
	}
	return result, nil
}

func (repository *Workouts) Complete(ctx context.Context, userID, workoutID string) (workout.Summary, error) {
	tx, err := repository.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return workout.Summary{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	result, err := getWorkoutForUpdate(ctx, tx, userID, workoutID)
	if errors.Is(err, pgx.ErrNoRows) {
		return workout.Summary{}, workout.ErrNotFound
	}
	if err != nil {
		return workout.Summary{}, err
	}
	if result.Status == 2 {
		if err := tx.Commit(ctx); err != nil {
			return workout.Summary{}, err
		}
		return result, nil
	}
	if result.Status != 1 {
		return workout.Summary{}, workout.ErrTerminalState
	}

	err = tx.QueryRow(ctx, `
		UPDATE workout_sessions
		SET status = 2,
			completed_at = now(),
			duration_seconds = GREATEST(0, EXTRACT(EPOCH FROM (now() - started_at))::integer),
			completed_exercise_count = 0,
			completed_set_count = 0,
			total_reps = 0,
			total_volume_kg = 0,
			updated_at = now()
		WHERE id = $1 AND user_id = $2 AND status = 1
		RETURNING id, status, started_at, completed_at, duration_seconds,
			completed_exercise_count, completed_set_count, total_reps,
			total_volume_kg::text
	`, workoutID, userID).Scan(summaryScanTargets(&result)...)
	if err != nil {
		return workout.Summary{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return workout.Summary{}, err
	}
	return result, nil
}

func (repository *Workouts) GetSummary(ctx context.Context, userID, workoutID string) (workout.Summary, error) {
	result := workout.Summary{}
	err := repository.pool.QueryRow(ctx, `
		SELECT id, status, started_at, completed_at, duration_seconds,
			completed_exercise_count, completed_set_count, total_reps,
			total_volume_kg::text
		FROM workout_sessions
		WHERE id = $1 AND user_id = $2 AND status = 2
	`, workoutID, userID).Scan(summaryScanTargets(&result)...)
	if errors.Is(err, pgx.ErrNoRows) {
		return workout.Summary{}, workout.ErrNotFound
	}
	return result, err
}

func getActiveWorkout(ctx context.Context, tx pgx.Tx, userID string) (workout.Summary, error) {
	result := workout.Summary{}
	err := tx.QueryRow(ctx, `
		SELECT id, status, started_at
		FROM workout_sessions
		WHERE user_id = $1 AND status = 1
	`, userID).Scan(&result.ID, &result.Status, &result.StartedAt)
	return result, err
}

func getWorkoutForUpdate(ctx context.Context, tx pgx.Tx, userID, workoutID string) (workout.Summary, error) {
	result := workout.Summary{}
	err := tx.QueryRow(ctx, `
		SELECT id, status, started_at, completed_at, duration_seconds,
			completed_exercise_count, completed_set_count, total_reps,
			total_volume_kg::text
		FROM workout_sessions
		WHERE id = $1 AND user_id = $2
		FOR UPDATE
	`, workoutID, userID).Scan(summaryScanTargets(&result)...)
	return result, err
}

func summaryScanTargets(result *workout.Summary) []any {
	return []any{
		&result.ID,
		&result.Status,
		&result.StartedAt,
		&result.CompletedAt,
		&result.DurationSeconds,
		&result.CompletedExerciseCount,
		&result.CompletedSetCount,
		&result.TotalReps,
		&result.TotalVolumeKG,
	}
}

var _ workout.Repository = (*Workouts)(nil)
