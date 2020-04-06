package main

import (
	"context"
	"log"
	"time"

	"github.com/danoand/gotomate-api/config"
	"github.com/danoand/gotomate-api/routes"
	"github.com/danoand/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {

	// Connect to the MongoDB database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(
		ctx,
		options.Client().ApplyURI(config.DbCredentials()["url"]))
	if err != nil {
		log.Fatal("FATAL: %v - fatal error connecting to the MongoDB database. See: %v", utils.FileLine(), err)
	}

	// Stand up the gin based server
	gin.SetMode(gin.TestMode)
	router := routes.SetupRouter()

	log.Printf("INFO: %v - start up the web server\n", utils.FileLine())
	router.Run()
}
