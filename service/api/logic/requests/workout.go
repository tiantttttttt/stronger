package requests

type CreateWorkout struct{}

type UpdateWorkout struct {
	Status int16 `json:"status" binding:"required" enums:"2"` // 当前仅支持更新为已完成。
}
