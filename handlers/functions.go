package handlers

import (
	"context"
	"log"
	"time"

	"github.com/danoand/rpachain-api/models"
	"github.com/danoand/utils"
	"github.com/globalsign/mgo/bson"
)

// logblockwrite logs block write data to a data store
func (hlr *HandlerEnv) logblockwrite(
	custid string,
	reqid string,
	mnfst models.Manifest,
	hsh string,
	txn map[string]string,
	custref map[string]interface{}) {

	var (
		err error
		obj models.BlockWrite
	)

	// Construct the object to be logged
	obj.ID = bson.NewObjectId().Hex()
	obj.CustomerID = custid
	obj.ChainNetwork = hlr.GoChainNetworkString
	obj.RequestID = reqid
	obj.TimeStamp = time.Now().In(hlr.TimeLocationCT).Format(time.RFC3339)
	obj.RequestID = mnfst.RequestID
	obj.Manifest = mnfst
	obj.TransactionHash = hsh
	obj.BlockTransaction = txn

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
