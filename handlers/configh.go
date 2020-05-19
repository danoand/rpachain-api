package handlers

import (
	"time"

	"github.com/gochain/web3"
	"github.com/minio/minio-go"
	che "github.com/patrickmn/go-cache"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"

	fak "github.com/contribsys/faktory/client"
	wrk "github.com/contribsys/faktory_worker_go"
)

// HandlerEnv houses config data needed for route handler execution
type HandlerEnv struct {
	TimeLocationCT         *time.Location
	Client                 *mongo.Client
	Database               *mongo.Database
	GridFS                 *gridfs.Bucket
	CollStatus             *mongo.Collection
	CollBlockWrites        *mongo.Collection
	CollAccounts           *mongo.Collection
	GoChainNetwork         web3.Client
	GoChainNetworkString   string
	GoChainCntrAddrLogHash string        // address of contract
	GoChainCntrABIURL      string        // web access to smart contract ABI
	SpacesClient           *minio.Client //
	FaktoryManager         *wrk.Manager  // Faktory worker queue manager (work queued jobs)
	FaktoryClient          *fak.Client   // Faktory worker client (queue jobs)
	Cache                  *che.Cache
}
