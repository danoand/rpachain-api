package handlers

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	fak "github.com/contribsys/faktory/client"
	bsn "github.com/globalsign/mgo/bson"
	"github.com/gochain/web3"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/danoand/rpachain-api/models"
	"github.com/danoand/utils"
)

// logblockwrite logs block write data to a data store
func (hlr *HandlerEnv) logblockwrite(
	custid string,
	source string,
	reqid string,
	mnfst models.Manifest,
	hsh string,
	txnhsh string,
	txn map[string]string,
	funktion string,
	mtx *sync.Mutex) {

	var (
		err error
		obj models.BlockWrite
	)

	// Lock the mutex so that a paired saveTxnReceipt goroutine will wait
	mtx.Lock()
	defer mtx.Unlock()

	// Construct the object to be logged
	obj.ID = bsn.NewObjectId().Hex()
	obj.CustomerID = custid
	obj.Source = source
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

// saveTxnReceipt waits for the receipt of a blockchain transaction and writes pertinent data
//    to the database
func (hlr *HandlerEnv) saveTxnReceipt(
	txnSC *web3.Transaction,
	reqID string,
	mtx *sync.Mutex) {

	var bwrt models.BlockWrite

	// Lock the mutex so that a paired saveTxnReceipt goroutine will wait
	mtx.Lock()
	defer mtx.Unlock()

	// Fetch the transaction receipt (will wait if the transaction is still pending)
	ctx, cncl := context.WithTimeout(context.Background(), 30*time.Second)
	defer cncl()

	// Wait for the transaction receipt
	rcpt, err := web3.WaitForReceipt(ctx, hlr.GoChainNetwork, txnSC.Hash)
	if err != nil {
		// error occurred fetching a transaction receipt
		log.Printf("ERROR: %v - error occurred fetching a transaction receipt for request id: %v. See: %v\n",
			utils.FileLine(),
			reqID,
			err)

		return
	}

	// Fetch the blockwrite log object from the database
	ctxmg1, cnclmg1 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cnclmg1()

	err = hlr.CollBlockWrites.FindOne(ctxmg1, bson.M{"requestid": reqID}).Decode(&bwrt)
	if err != nil {
		// error fetching a blockwrite object from the database
		log.Printf("ERROR: %v - error fetching a blockwrite object from the database for request id: %v. See: %v\n",
			utils.FileLine(),
			reqID,
			err)

		return
	}

	// Update the blockwrite object back in the database
	ctxmg2, cnclmg2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cnclmg2()
	filter := bson.D{{"requestid", reqID}}
	update := bson.D{
		{"$set", bson.D{
			{"blocknumber", fmt.Sprintf("%d", rcpt.BlockNumber)},
			{"blockhash", fmt.Sprintf("0x%x", rcpt.BlockHash)},
			{"transactionlog", rcpt.ParsedLogs}}}}
	// Update the blockwrite object in the database
	_, err = hlr.CollBlockWrites.UpdateOne(
		ctxmg2,
		filter,
		update)
	if err != nil {
		// error updating the block
		log.Printf("ERROR: %v - error fetching a blockwrite object from the database for request id: %v. See: %v\n",
			utils.FileLine(),
			reqID,
			err)

		return
	}

	log.Printf("INFO: %v - updated write object with request id: %v with block number, hash, and log\n",
		utils.FileLine(),
		reqID)

	return
}
