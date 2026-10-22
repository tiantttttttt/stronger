package handler

import (
	"context"
	standardhttp "net/http"
	"time"

	"github.com/gin-gonic/gin"

	"stronger/server/service/api/logic"
)

// @Summary 健康检查
// @Tags health
// @Success 200 {object} responses.Health
// @Failure 503 {object} responses.Health
// @Router /health [get]
func Health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	if err := logic.HealthLogic.Ping(ctx); err != nil {
		c.JSON(standardhttp.StatusServiceUnavailable, gin.H{
			"status":   "error",
			"database": "unavailable",
			"code":     "database_unavailable",
		})
		return
	}

	c.JSON(standardhttp.StatusOK, gin.H{
		"status":   "ok",
		"database": "ok",
	})
}
