package handlers

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
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

// GetGinContextValStr returns a gin context value as a string
func GetGinContextValStr(c *gin.Context, key string) (string, error) {
	var (
		ok     bool
		tmpVal interface{}
		retStr string
	)

	// Empty key value?
	if len(key) == 0 {
		// missing key value
		return "", fmt.Errorf("missing key value")
	}

	// Grab the key value from the gin context
	tmpVal, ok = c.Get(key)
	if !ok {
		// error occurred getting a gin context value
		return "", fmt.Errorf("error getting the gin context value for key: %v", key)
	}

	// Assert the value as a string
	retStr, ok = tmpVal.(string)
	if !ok {
		// failed the type assertion
		return "", fmt.Errorf("key value: %v was not typed as a string", key)
	}

	return retStr, nil
}
