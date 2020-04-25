package handlers

import (
	"context"
	"log"
	"time"

	"github.com/danoand/rpachain-api/config"
	"github.com/danoand/rpachain-api/hash"
	"github.com/danoand/rpachain-api/models"
	"github.com/danoand/utils"
)

// logblockwrite logs block write data to a data store
func (hlr *HandlerEnv) logblockwrite(
	hsh string, mnfst hash.Manifest, txn map[string]string, ref string) {
	var (
		err error
		obj models.BlockWrite
	)

	obj.ID = bson.NewObjectID().Hex()
	obj.CustomerID = CustomerID
	obj.ChainNetwork = hlr.GoChainNetworkString
	obj.RequestID = reqid
	obj.TimeStamp = time.Now().In(config.Consts["timezone"]).Format(time.RFC3339)

	obj.RequestID = mnfst.ID
	obj.TransactionHash = hsh
	obj.Manifest = mnfst
	obj.BlockTransaction = txn
	obj.CustomerReference = ref

	// Insert log document into the database
	_, err = hlr.CollBlockWrites.InsertOne(context.TODO(), obj)
	if err != nil {
		// error writing an object to the database
		log.Printf("ERROR: %v - error writing an object to the database. See: %v\n",
			utils.FileLine(),
			err)
	}

	return
}
