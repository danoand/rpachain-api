package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/danoand/gotomate-api/handlers"
)

var router *gin.Engine

// initRoutes initializes API routes
func initRoutes() {

	apiv1 := router.Group("/api/v1")
	{
		apiv1.GET("/status", handlers.Status)
	}
}

// SetupRouter creates a router for use downstream
func SetupRouter() *gin.Engine {
	router = gin.Default()
	initRoutes()

	return router
}
