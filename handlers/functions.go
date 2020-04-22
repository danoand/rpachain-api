package handlers

import (
	"context"
	"log"

	"github.com/danoand/rpachain-api/hash"
	"github.com/danoand/rpachain-api/models"
	"github.com/danoand/utils"
	"github.com/globalsign/mgo/bson"
	"github.com/gochain/web3"
)

// logblockwrite logs block write data to a data store
func (hlr *HandlerEnv) logblockwrite(hsh string, mnfst hash.Manifest, txn *web3.Transaction, ref string) {
	var (
		err error
		obj models.BlockWrite
	)

	obj.ID = bson.NewObjectId()
	obj.Hash = hsh
	obj.Manifest = mnfst
	obj.BlockTransaction = *txn
	obj.BlockDataRef = ref

	_, err = hlr.CollBlockWrites.InsertOne(context.TODO(), obj)
	if err != nil {
		// error writing an object to the database
		log.Printf("ERROR: %v - error writing an object to the database. See: %v\n",
			utils.FileLine(),
			err)
	}

	return
}
