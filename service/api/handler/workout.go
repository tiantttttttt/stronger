package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"stronger/server/service/api/logic"
	"stronger/server/service/api/logic/requests"
	"stronger/server/service/api/middleware"
	"stronger/server/service/model"
)

// Workouts 处理 HTTP 参数与响应，业务与事务由 Logic 负责。
type Workouts struct{}

var WorkoutsHandler = &Workouts{}

// @Summary 创建空训练
// @Tags workouts
// @Security BearerAuth
// @Param body body requests.CreateWorkout false "创建请求"
// @Success 200 {object} responses.WorkoutRecord
// @Router /api/v1/workouts [post]
func (h Workouts) Create(c *gin.Context) {
	var request requests.CreateWorkout
	if err := c.ShouldBindJSON(&request); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"code": "invalid_request", "message": "request body must be valid JSON"})
		return
	}
	userID, _ := middleware.CurrentUserID(c)
	record, err := logic.WorkoutsLogic.Create(c.Request.Context(), userID)
	writeWorkoutResult(c, record, err)
}

// @Summary 更新训练状态
// @Tags workouts
// @Security BearerAuth
// @Param id path string true "训练 UUID" format(uuid)
// @Param body body requests.UpdateWorkout true "更新请求"
// @Success 200 {object} responses.WorkoutRecord
// @Router /api/v1/workouts/{id} [put]
func (h Workouts) Update(c *gin.Context) {
	workoutID, ok := UUIDPathParam(c, "id")
	if !ok {
		return
	}
	var request requests.UpdateWorkout
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "invalid_request", "message": "request body must contain a valid status"})
		return
	}
	userID, _ := middleware.CurrentUserID(c)
	record, err := logic.WorkoutsLogic.Update(c.Request.Context(), userID, workoutID.String(), request)
	writeWorkoutResult(c, record, err)
}

// @Summary 查询已完成训练统计
// @Tags workouts
// @Security BearerAuth
// @Param id path string true "训练 UUID" format(uuid)
// @Success 200 {object} responses.WorkoutStat
// @Router /api/v1/workouts/{id}/stat [get]
func (h Workouts) Stat(c *gin.Context) {
	workoutID, ok := UUIDPathParam(c, "id")
	if !ok {
		return
	}
	userID, _ := middleware.CurrentUserID(c)
	stat, err := logic.WorkoutsLogic.Stat(c.Request.Context(), userID, workoutID.String())
	writeWorkoutResult(c, stat, err)
}

// writeWorkoutResult 共用响应写入与训练错误映射，不限定响应数据类型。
func writeWorkoutResult[T any](c *gin.Context, data T, err error) {
	switch {
	case err == nil:
		c.JSON(http.StatusOK, data)
	case errors.Is(err, model.ErrInvalidRequest):
		c.JSON(http.StatusBadRequest, gin.H{"code": "invalid_request", "message": "status must be 2 (completed)"})
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
