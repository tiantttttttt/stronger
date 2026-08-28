package http

import (
	"context"
	standardhttp "net/http"
	"time"

	"github.com/gin-gonic/gin"

	"stronger/server/internal/auth"
	"stronger/server/internal/workout"
)

// DatabasePinger is the health-check subset implemented by pgxpool.Pool.
type DatabasePinger interface {
	Ping(context.Context) error
}

// Dependencies contains the router's infrastructure and runtime policy.
type Dependencies struct {
	Database    DatabasePinger
	Environment string
	Tokens      *auth.TokenManager
	DevUsers    auth.DevUserEnsurer
	Workouts    workout.Repository
}

// NewRouter builds the API router.
func NewRouter(dependencies Dependencies) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.GET("/health", healthHandler(dependencies.Database))
	api := router.Group("/api/v1")
	auth.RegisterRoutes(api, dependencies.Environment, dependencies.Tokens, dependencies.DevUsers)
	if dependencies.Workouts != nil {
		workout.RegisterRoutes(api, dependencies.Tokens, dependencies.Workouts)
	}
	return router
}

func healthHandler(database DatabasePinger) gin.HandlerFunc {
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
