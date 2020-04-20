package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/danoand/gotomate-api/config"
	"github.com/danoand/gotomate-api/handlers"
	"github.com/danoand/gotomate-api/routes"
	"github.com/danoand/utils"
	"github.com/gin-gonic/gin"
	"github.com/gochain/web3"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var err error
var hndlr handlers.HandlerEnv

func main() {

	// Construct the MongoDB connection string
	connStr := fmt.Sprintf(config.Cfg.MGDBURLString, config.Cfg.MGDBPassword)
	log.Printf("DEBUG: the db connection string is: %v\n", connStr)

	// Connect to the MongoDB database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	hndlr.Client, err = mongo.Connect(
		ctx,
		options.Client().ApplyURI(connStr))
	if err != nil {
		log.Fatalf("FATAL: %v - fatal error connecting to the MongoDB database. See: %v", utils.FileLine(), err)
	}

	tmpctx, tmpcancel := context.WithTimeout(context.Background(), 5*time.Second)
	err = hndlr.Client.Ping(tmpctx, nil)
	if err != nil {
		log.Fatalf("FATAL: %v - fatal error pinging the MongoDB database. See: %v", utils.FileLine(), err)
	}
	tmpcancel()

	// Configure handler object
	hndlr.TimeLocationCT, err = time.LoadLocation("America/Chicago")
	if err != nil {
		// error loading a timezone
		log.Fatalf("FATAL: %v - error loading a timezone. See: %v\n", utils.FileLine(), err)
	}
	hndlr.Database = hndlr.Client.Database("gotomate-dev")
	hndlr.CollStatus = hndlr.Database.Collection("status")
	// Create an object dialing the GoChain network
	hndlr.GoChainNetwork, err = web3.Dial(config.Cfg.GoChainURL)
	if err != nil {
		// error dialing the GoChain network/blockchain
		log.Printf("ERROR: %v - error dialing the GoChain testnet network/blockchain. See: %v\n",
			utils.FileLine(),
			err)

		os.Exit(1)
	}

	// Stand up the gin based server
	gin.SetMode(gin.TestMode)
	router := routes.SetupRouter(&hndlr)

	log.Printf("INFO: %v - start up the web server on localhost:8080\n", utils.FileLine())
	router.Run("localhost:8080")
}
