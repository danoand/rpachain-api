package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	fak "github.com/contribsys/faktory/client"
	wrk "github.com/contribsys/faktory_worker_go"
	"github.com/danoand/rpachain-api/config"
	"github.com/danoand/rpachain-api/handlers"
	"github.com/danoand/rpachain-api/routes"
	"github.com/danoand/utils"
	"github.com/gin-gonic/gin"
	"github.com/gochain/web3"
	"github.com/minio/minio-go"
	che "github.com/patrickmn/go-cache"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var err error
var hndlr handlers.HandlerEnv
var workerOverride bool // command line argurment indicating this instance is a worker instance

func main() {

	// Starting this instance as a worker instanc (from the command line)
	if len(os.Args) > 1 {
		// we have passed arguments... check them out
		if os.Args[1] == "worker" {
			// starting this as a worker instance
			log.Printf("INFO: %v - starting this instance as a worker web service", utils.FileLine())
			workerOverride = true
		}
	}

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

	// If starting as a worker instance, override read environment variables
	if workerOverride {
		config.Cfg.WrkIsWorkerInstance = true
	}

	// Configure handler object
	hndlr.TimeLocationCT, err = time.LoadLocation(config.Consts["timezone"])
	if err != nil {
		// error loading a timezone
		log.Fatalf("FATAL: %v - error loading a timezone. See: %v\n", utils.FileLine(), err)
	}
	hndlr.Database = hndlr.Client.Database("rpachain-dev")
	hndlr.CollStatus = hndlr.Database.Collection("status")
	hndlr.CollBlockWrites = hndlr.Database.Collection("blockwrites")
	hndlr.CollAccounts = hndlr.Database.Collection("accounts")
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
	// Declare a cache
	hndlr.Cache = che.New(60*time.Minute, 70*time.Minute)

	// Set the localhost port for web api mode
	port := config.Consts["localport_api"]

	// *******************************
	// Configure Faktory worker instance information (so this instance can excute worker jobs)
	if config.Cfg.WrkIsWorkerInstance {
		// this execution app instance is a worker instance
		// register worker functions (job handlers)
		log.Printf("INFO: %v - setting up Faktory worker functions\n", utils.FileLine())
		// Override the localhost port for worker mode
		port = config.Consts["localport_wrk"]

		hndlr.FaktoryManager = wrk.NewManager()

		hndlr.FaktoryManager.Register("TestFunktion", hndlr.TestFunktion)
		hndlr.FaktoryManager.Register("CreateHashTarBallFile", hndlr.ZipRequest)

		hndlr.FaktoryManager.Concurrency = 10
		hndlr.FaktoryManager.ProcessStrictPriorityQueues("rpa_high", "rpa_default")

		log.Printf("INFO: %v - starting Faktory job processing\n", utils.FileLine())
		go hndlr.FaktoryManager.Run() // start listener for Faktory jobs in a go routine
	}

	// Configure Facktory client instance information (so this instance can queue worker jobs)
	if !config.Cfg.WrkIsWorkerInstance {
		// this execution app instance is a web instance (and will queue workers)
		log.Printf("INFO: %v - setting up Faktory queuing instance (in web API mode)\n", utils.FileLine())

		hndlr.FaktoryClient, err = fak.Open()
		if err != nil {
			// error creating a Faktory client
			log.Fatalf("ERROR: %v - error creating a Faktory client. See: %v\n",
				utils.FileLine(),
				err)
		}
	}

	// In the Heroku environment?
	if config.Cfg.IsHerokuEnv {
		// Override the localhost port for execution on Heroku
		port = fmt.Sprintf(":%v", os.Getenv("PORT"))
	}

	// Stand up the gin based server
	gin.SetMode(gin.DebugMode)
	router := routes.SetupRouter(&hndlr)

	log.Printf("INFO: %v - start up the web server on: %v\n", utils.FileLine(), port)
	router.Run(port)
}
