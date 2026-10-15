package handler

import (
	"context"
	standardhttp "net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type DatabasePinger interface{ Ping(context.Context) error }

// Health 检查数据库连接。
// @Summary 健康检查
// @Tags health
// @Produce json
// @Success 200 {object} responses.Health
// @Failure 503 {object} responses.Health
// @Router /health [get]
func Health(database DatabasePinger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		if err := database.Ping(ctx); err != nil {
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
}
