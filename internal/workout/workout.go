package workout

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"stronger/server/internal/auth"
	"stronger/server/internal/platform/httpx"
)

var (
	ErrNotFound          = errors.New("workout not found")
	ErrProfileIncomplete = errors.New("profile incomplete")
	ErrTerminalState     = errors.New("workout is already in a terminal state")
)

type Summary struct {
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

type Repository interface {
	Start(context.Context, string) (Summary, error)
	Complete(context.Context, string, string) (Summary, error)
	GetSummary(context.Context, string, string) (Summary, error)
}

func RegisterRoutes(router *gin.RouterGroup, tokens *auth.TokenManager, repository Repository) {
	workouts := router.Group("/workouts", tokens.Middleware())
	workouts.POST("", func(c *gin.Context) {
		userID, _ := auth.CurrentUserID(c)
		summary, err := repository.Start(c.Request.Context(), userID)
		writeResult(c, summary, err, http.StatusOK)
	})
	workouts.POST("/:id/complete", func(c *gin.Context) {
		workoutID, ok := httpx.UUIDPathParam(c, "id")
		if !ok {
			return
		}
		userID, _ := auth.CurrentUserID(c)
		summary, err := repository.Complete(c.Request.Context(), userID, workoutID.String())
		writeResult(c, summary, err, http.StatusOK)
	})
	workouts.GET("/:id/summary", func(c *gin.Context) {
		workoutID, ok := httpx.UUIDPathParam(c, "id")
		if !ok {
			return
		}
		userID, _ := auth.CurrentUserID(c)
		summary, err := repository.GetSummary(c.Request.Context(), userID, workoutID.String())
		writeResult(c, summary, err, http.StatusOK)
	})
}

func writeResult(c *gin.Context, summary Summary, err error, successStatus int) {
	switch {
	case err == nil:
		c.JSON(successStatus, summary)
	case errors.Is(err, ErrProfileIncomplete):
		c.JSON(http.StatusConflict, gin.H{"code": "profile_incomplete", "message": "complete the required profile before starting a workout"})
	case errors.Is(err, ErrTerminalState):
		c.JSON(http.StatusConflict, gin.H{"code": "workout_terminal", "message": "workout can no longer be modified"})
	case errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"code": "workout_not_found", "message": "workout not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"code": "internal_error", "message": "request could not be completed"})
	}
}
