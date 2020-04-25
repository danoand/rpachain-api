package handlers

import (
	"time"

	"github.com/gochain/web3"
	"github.com/minio/minio-go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
)

// HandlerEnv houses config data needed for route handler execution
type HandlerEnv struct {
	TimeLocationCT       *time.Location
	Client               *mongo.Client
	Database             *mongo.Database
	GridFS               *gridfs.Bucket
	CollStatus           *mongo.Collection
	CollBlockWrites      *mongo.Collection
	GoChainNetwork       web3.Client
	GoChainNetworkString string
	SpacesClient         *minio.Client
}
