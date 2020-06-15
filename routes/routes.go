package routes

import (
	"log"

	"github.com/gin-gonic/gin"
	ginsession "github.com/go-session/gin-session"

	"github.com/danoand/rpachain-api/config"
	"github.com/danoand/rpachain-api/handlers"
	hdl "github.com/danoand/rpachain-api/handlers"
	"github.com/danoand/utils"
)

var router *gin.Engine

// initStandardRoutes initializes API and related routes enabling the standard web application
func initStandardRoutes(hndlr *hdl.HandlerEnv) {

	web := router.Group("/webapp")
	web.Use(ginsession.New())
	{
		web.Static("/fonts", "webapp/app/fonts")
		web.Static("/styles", "webapp/app/styles")
		web.Static("/scripts", "webapp/app/scripts")
		web.Static("/app_components", "webapp/app_components")
		web.Static("/views", "webapp/app/views")
		web.Static("/images", "webapp/app/images")
		web.StaticFile("/", "webapp/app/index.html")

		web.GET("/getblockwrites", hndlr.WebAuth(), hndlr.GetBlockWrites)

		web.POST("/login", hndlr.Login)
		web.POST("/logoff", hndlr.Logoff)
	}

	apiv1 := router.Group("/api/v1")
	{
		apiv1.GET("/status", hndlr.Status)
		apiv1.POST("/blockwrite", hndlr.APIAuth(), hndlr.BlockWrite)
		apiv1.POST("/blockwritefiles", hndlr.APIAuth(), hndlr.BlockWriteFiles)
	}
}

// initWorkerRoutes initializes API and related routes enabling the standard web application
func initWorkerRoutes(hndlr *handlers.HandlerEnv) {

	apiv1 := router.Group("/wrk/v1")
	{
		apiv1.GET("/status", hndlr.FaktoryStatus)
	}
}

// SetupRouter creates a router for use downstream
func SetupRouter(hndlr *handlers.HandlerEnv) *gin.Engine {

	router = gin.Default()

	// Standard app instance? (ie. not a Faktory instance)
	if !config.Cfg.WrkIsWorkerInstance {
		// not a worker instance... configure the standard web service routes
		log.Printf("INFO: %v - setting up api web service routes\n", utils.FileLine())
		initStandardRoutes(hndlr)
	}

	// Worker app instance? (ie. is a Faktory instance)
	if config.Cfg.WrkIsWorkerInstance {
		// is a worker instance... configure the worker web service routes
		log.Printf("WRKR: %v - setting up worker web service routes\n", utils.FileLine())
		initWorkerRoutes(hndlr)
	}

	return router
}
