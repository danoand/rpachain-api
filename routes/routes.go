package routes

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/danoand/rpachain-api/config"
	"github.com/danoand/rpachain-api/handlers"
	mdw "github.com/danoand/rpachain-api/middleware"
	"github.com/danoand/utils"
)

var router *gin.Engine

// initStandardRoutes initializes API and related routes enabling the standard web application
func initStandardRoutes(hndlr *handlers.HandlerEnv) {

	apiv1 := router.Group("/api/v1")
	{
		apiv1.GET("/status", hndlr.Status)
		apiv1.POST("/blockwrite", mdw.APIAuth(), hndlr.BlockWrite)
		apiv1.POST("/blockwritefiles", mdw.APIAuth(), hndlr.BlockWriteFiles)
	}
}

// SetupRouter creates a router for use downstream
func SetupRouter(hndlr *handlers.HandlerEnv) *gin.Engine {
	router = gin.Default()

	// Standard app instance? (ie. not a Faktory instance)
	if !config.Cfg.WrkIsWorkerInstance {
		// not a worker instance... configure the standard web service routes
		log.Printf("INFO: %v - setting up standard web service routes\n", utils.FileLine())
		initStandardRoutes(hndlr)
	}

	return router
}
