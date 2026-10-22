package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"stronger/server/service/api/logic"
	"stronger/server/service/api/logic/responses"
	"stronger/server/service/api/middleware"
	"stronger/server/service/model"
)

// Workouts 处理 HTTP 参数与响应，业务与事务由 Logic 负责。
type Workouts struct{}

var WorkoutsHandler = &Workouts{}

// @Summary 开始空训练
// @Tags workouts
// @Security BearerAuth
// @Success 200 {object} responses.WorkoutSummary
// @Router /api/v1/workouts [post]
// @Failure 401,404,409,500 {object} responses.Error
func (h Workouts) Start(c *gin.Context) {
	userID, _ := middleware.CurrentUserID(c)
	summary, err := logic.WorkoutsLogic.Start(c.Request.Context(), userID)
	writeResult(c, summary, err, http.StatusOK)
}

// @Summary 完成训练
// @Tags workouts
// @Security BearerAuth
// @Param id path string true "训练 UUID" format(uuid)
// @Success 200 {object} responses.WorkoutSummary
// @Router /api/v1/workouts/{id}/complete [post]
// @Failure 400,401,404,409,500 {object} responses.Error
func (h Workouts) Complete(c *gin.Context) {
	workoutID, ok := UUIDPathParam(c, "id")
	if !ok {
		return
	}
	userID, _ := middleware.CurrentUserID(c)
	summary, err := logic.WorkoutsLogic.Complete(c.Request.Context(), userID, workoutID.String())
	writeResult(c, summary, err, http.StatusOK)
}

// @Summary 查询已完成训练总结
// @Tags workouts
// @Security BearerAuth
// @Param id path string true "训练 UUID" format(uuid)
// @Success 200 {object} responses.WorkoutSummary
// @Router /api/v1/workouts/{id}/summary [get]
// @Failure 400,401,404,500 {object} responses.Error
func (h Workouts) GetSummary(c *gin.Context) {
	workoutID, ok := UUIDPathParam(c, "id")
	if !ok {
		return
	}
	userID, _ := middleware.CurrentUserID(c)
	summary, err := logic.WorkoutsLogic.GetSummary(c.Request.Context(), userID, workoutID.String())
	writeResult(c, summary, err, http.StatusOK)
}

func writeResult(c *gin.Context, summary responses.WorkoutSummary, err error, successStatus int) {
	switch {
	case err == nil:
		c.JSON(successStatus, summary)
	case errors.Is(err, model.ErrProfileIncomplete):
		c.JSON(http.StatusConflict, gin.H{"code": "profile_incomplete", "message": "complete the required profile before starting a workout"})
	case errors.Is(err, model.ErrTerminalState):
		c.JSON(http.StatusConflict, gin.H{"code": "workout_terminal", "message": "workout can no longer be modified"})
	case errors.Is(err, model.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"code": "workout_not_found", "message": "workout not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"code": "internal_error", "message": "request could not be completed"})
	}
}
