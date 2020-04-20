package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/danoand/gotomate-api/handlers"
)

var router *gin.Engine

// initRoutes initializes API routes
func initRoutes(hndlr *handlers.HandlerEnv) {

	apiv1 := router.Group("/api/v1")
	{
		apiv1.GET("/status", hndlr.Status)
		apiv1.POST("/blockwrite", hndlr.BlockWrite)
	}
}

// SetupRouter creates a router for use downstream
func SetupRouter(hndlr *handlers.HandlerEnv) *gin.Engine {
	router = gin.Default()
	initRoutes(hndlr)

	return router
}
