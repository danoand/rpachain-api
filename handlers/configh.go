package handlers

import (
	"time"

	"github.com/gochain/web3"
	"github.com/minio/minio-go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"

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
	GoChainNetwork         web3.Client
	GoChainNetworkString   string
	GoChainCntrAddrLogHash string        // address of contract
	GoChainCntrABIURL      string        // web access to smart contract ABI
	SpacesClient           *minio.Client //
	FaktoryClient          *wrk.Manager  // Faktory client manager
}
