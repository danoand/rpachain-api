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
	txnhsh string,
	txn map[string]string,
	funktion string) {

	var (
		err error
		obj models.BlockWrite
	)

	// Construct the object to be logged
	obj.ID = bson.NewObjectId().Hex()
	obj.CustomerID = custid
	obj.ChainNetwork = hlr.GoChainNetworkString
	obj.ContractAddress = hlr.GoChainCntrAddrLogHash
	obj.RequestID = reqid
	obj.TimeStamp = time.Now().In(hlr.TimeLocationCT).Format(time.RFC3339)
	obj.RequestID = mnfst.RequestID
	obj.Manifest = mnfst
	obj.ManifestHash = hsh
	obj.TransactionHash = txnhsh
	obj.BlockTransaction = txn
	obj.Function = funktion

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
