package router

import (
	"github.com/gin-gonic/gin"

	"stronger/server/service/api/handler"
	"stronger/server/service/api/middleware"
)

// NewRouter 注册路由及鉴权，数据库由服务启动流程初始化。
func NewRouter(debug bool, tokens *middleware.TokenManager) *gin.Engine {
	if debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(gin.Recovery())
	router.GET("/health", handler.Health)
	api := router.Group("/api/v1")
	auth := handler.Auth{Debug: debug, Tokens: tokens}
	if debug {
		api.POST("/auth/dev/login", auth.DevLogin)
	}
	api.GET("/me", tokens.Middleware(), auth.Me)
	routes := api.Group("/workouts", tokens.Middleware())
	routes.POST("", handler.WorkoutsHandler.Create)
	routes.PUT("/:id", handler.WorkoutsHandler.Update)
	routes.GET("/:id/stat", handler.WorkoutsHandler.Stat)
	return router
}
