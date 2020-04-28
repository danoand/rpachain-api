package handlers

import (
	"context"
	"log"
	"time"

	fak "github.com/contribsys/faktory/client"

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

	// Queue a job to create a zip ball
	job := fak.NewJob("CreateHashTarBallFile", mnfst.RequestID)
	job.Queue = "rpa_high"
	job.Custom = map[string]interface{}{"for_requestid": mnfst.RequestID}
	job.Retry = -1 // don't retry job if it fails

	// Push the job to the Faktory instance
	err = hlr.FaktoryClient.Push(job)
	if err != nil {
		// error queuing a job to zip files for a particular request
		log.Printf("ERROR: %v - error queuing a job to zip files for a particular request: %v. See: %v\n",
			utils.FileLine(),
			mnfst.RequestID,
			err)
	}
	if err == nil {
		// log job queue information
		log.Printf("INFO: %v - queued zip job: %v [queue: %v] for request: %v\n",
			utils.FileLine(),
			job.Jid,
			job.Queue,
			mnfst.RequestID)
	}

	return
}
