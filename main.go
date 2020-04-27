package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	wrk "github.com/contribsys/faktory_worker_go"
	"github.com/danoand/rpachain-api/config"
	"github.com/danoand/rpachain-api/handlers"
	"github.com/danoand/rpachain-api/routes"
	"github.com/danoand/utils"
	"github.com/gin-gonic/gin"
	"github.com/gochain/web3"
	"github.com/minio/minio-go"
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
	hndlr.TimeLocationCT, err = time.LoadLocation(config.Consts["timezone"])
	if err != nil {
		// error loading a timezone
		log.Fatalf("FATAL: %v - error loading a timezone. See: %v\n", utils.FileLine(), err)
	}
	hndlr.Database = hndlr.Client.Database("rpachain-dev")
	hndlr.CollStatus = hndlr.Database.Collection("status")
	hndlr.CollBlockWrites = hndlr.Database.Collection("blockwrites")
	// Create an object dialing the GoChain network
	hndlr.GoChainNetwork, err = web3.Dial(config.Cfg.GoChainURL)
	if err != nil {
		// error dialing the GoChain network/blockchain
		log.Printf("ERROR: %v - error dialing the GoChain testnet network/blockchain. See: %v\n",
			utils.FileLine(),
			err)

		os.Exit(1)
	}
	hndlr.GoChainNetworkString = config.Cfg.GoChainURL
	// Assign contract address from environment variable
	hndlr.GoChainCntrAddrLogHash = config.Cfg.GoCntrtLogAddr
	// Assign contract ABI URL (web access)
	hndlr.GoChainCntrABIURL = config.Cfg.GoCntrtABIURL
	// Create a client object referencing the Spaces instance
	hndlr.SpacesClient, err = minio.New(
		"nyc3.digitaloceanspaces.com",
		config.Cfg.SpacesAccessKey,
		config.Cfg.SpacesSecretKey,
		true)
	if err != nil {
		// error establishing a client referencing the Spaces instance
		log.Printf("ERROR: %v - error establishing a client referencing the Spaces instance. See: %v\n",
			utils.FileLine(),
			err)

		os.Exit(1)
	}

	// *******************************
	// Configure worker instance information
	if config.Cfg.WrkIsWorkerInstance {
		// this execution app instance is a worker instance
		// register worker functions (job handlers)
		log.Printf("INFO: %v - setting up worker functions\n", utils.FileLine())

		hndlr.FaktoryClient = wrk.NewManager()
		hndlr.FaktoryClient.Register("TestFunktion", hndlr.TestFunktion)
		hndlr.FaktoryClient.Concurrency = 10
		hndlr.FaktoryClient.ProcessStrictPriorityQueues("rpa_high", "rpa_default")

		log.Printf("INFO: %v - starting Faktory job processing\n", utils.FileLine())
		go hndlr.FaktoryClient.Run() // start listener for Faktory jobs in a go routine
	}

	// Stand up the gin based server
	gin.SetMode(gin.TestMode)
	router := routes.SetupRouter(&hndlr)

	log.Printf("INFO: %v - start up the web server on localhost:8080\n", utils.FileLine())
	router.Run("localhost:8080")
}
