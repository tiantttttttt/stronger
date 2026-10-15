package router

import (
	"github.com/gin-gonic/gin"

	"stronger/server/service/api/handler"
	"stronger/server/service/api/middleware"
)

// Dependencies contains the router's infrastructure and runtime policy.
type Dependencies struct {
	Database handler.DatabasePinger
	Debug    bool
	Tokens   *middleware.TokenManager
	DevUsers handler.DevUserEnsurer
	Workouts handler.WorkoutLogic
}

// NewRouter builds the API router.
func NewRouter(dependencies Dependencies) *gin.Engine {
	if dependencies.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(gin.Recovery())
	router.GET("/health", handler.Health(dependencies.Database))
	api := router.Group("/api/v1")
	auth := handler.Auth{Debug: dependencies.Debug, Tokens: dependencies.Tokens, DevUsers: dependencies.DevUsers}
	if dependencies.Debug {
		api.POST("/auth/dev/login", auth.DevLogin)
	}
	api.GET("/me", dependencies.Tokens.Middleware(), auth.Me)
	if dependencies.Workouts != nil {
		workouts := handler.Workouts{Logic: dependencies.Workouts}
		routes := api.Group("/workouts", dependencies.Tokens.Middleware())
		routes.POST("", workouts.Start)
		routes.POST("/:id/complete", workouts.Complete)
		routes.GET("/:id/summary", workouts.GetSummary)
	}
	return router
}
