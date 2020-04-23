package routes

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/danoand/rpachain-api/config"
	"github.com/danoand/rpachain-api/handlers"
	"github.com/danoand/utils"
)

var router *gin.Engine

// initStandardRoutes initializes API and related routes enabling the standard web application
func initStandardRoutes(hndlr *handlers.HandlerEnv) {

	apiv1 := router.Group("/api/v1")
	{
		apiv1.GET("/status", hndlr.Status)
		apiv1.POST("/blockwrite", hndlr.BlockWrite)
		apiv1.POST("/blockwritefiles", hndlr.BlockWriteFiles)
	}
}

// SetupRouter creates a router for use downstream
func SetupRouter(hndlr *handlers.HandlerEnv) *gin.Engine {
	router = gin.Default()

	// Standard app instance? (ie. not a Faktory instance)
	if !config.Cfg.IsWorkerInstance {
		// not a worker instance... configure the standard web service routes
		log.Printf("INFO: %v - setting up standard web service routes\n", utils.FileLine())
		initStandardRoutes(hndlr)
	}

	if config.Cfg.IsWorkerInstance {
		// set up a worker instance... configure the worker routes
		log.Printf("INFO: %v - setting up standard web service routes\n", utils.FileLine())
		// TODO: add a function here to stand up worker routes
	}

	return router
}
