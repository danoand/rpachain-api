package handlers

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/danoand/rpachain-api/config"
	"github.com/danoand/rpachain-api/models"
	"github.com/globalsign/mgo/bson"
	"github.com/gochain/web3"
	"github.com/minio/minio-go"
	"lukechampine.com/blake3"

	"github.com/danoand/utils"
	"github.com/gin-gonic/gin"
)

// BlockWriteFiles hashes request files and writes the manifest hash to the blockchain
func (hlr *HandlerEnv) BlockWriteFiles(c *gin.Context) {
	var (
		err             error
		custid, custref string
		origin, event   string
		mnfst           models.Manifest
		// reqBytes []byte
		tmpInt = make(map[string]interface{})
		errMap = make(map[string]string)
		rspMap = make(map[string]interface{})
		bwMtx  sync.Mutex
	)

	custref = "N/A"

	mnfst.MetaData = make(map[string]interface{})

	// Get the customer document id from the gin context
	custid, err = GetGinContextValStr(c.Copy(), config.Consts["cxtRequestOrigin"])
	if err != nil {
		// error fetching the customer document id from the gin context
		log.Printf("ERROR: %v - error fetching the customer document id from the gin context. See: %v\n",
			utils.FileLine(),
			err)
		errMap["msg"] = "error processing the request"

		c.JSON(http.StatusInternalServerError, errMap)
		return
	}

	// Determine the origin of this request (web or api)
	origin, err = GetGinContextValStr(c.Copy(), config.Consts["cxtRequestOrigin"])
	if err != nil {
		// error fetching the request body
		log.Printf("ERROR: %v - error determining the origin of this inbound request. See: %v\n",
			utils.FileLine(),
			err)
		errMap["msg"] = "error processing the request"

		c.JSON(http.StatusInternalServerError, errMap)
		return
	}

	// Set up a Blake3 "hasher"
	blk3hshr := blake3.New(256, nil)

	// Parse the multi-part request
	form, err := c.MultipartForm()
	if err != nil {
		// error parsing the request
		log.Printf("ERROR: %v - error parsing the request. See: %v\n",
			utils.FileLine(),
			err)

		errMap["msg"] = "error parsing the request"
		c.JSON(http.StatusBadRequest, errMap)
		return
	}

	// Update the hash manifest
	mnfst.RequestID = bson.NewObjectId().Hex()
	mnfst.TimeStamp = time.Now().In(hlr.TimeLocationCT).Format(time.RFC3339)

	// Web origin request tasks
	if origin == config.Consts["web"] {

		// Grab the title value
		if len(form.Value["title"]) != 0 && len(form.Value["title"][0]) != 0 {
			// Get the title value
			tmpInt["title"] = form.Value["title"][0]
		}
		// Grab the text content value
		if len(form.Value["content"]) != 0 && len(form.Value["content"][0]) != 0 {
			// Get the title value
			tmpInt["content"] = form.Value["content"][0]
		}
		// Grab the meta data value
		if len(form.Value["meta_data_01"]) != 0 && len(form.Value["meta_data_01"][0]) != 0 {
			// Get the title value
			tmpInt["meta_data_01"] = form.Value["meta_data_01"][0]
		}
		// Assign the inbound data to the manifest
		if len(tmpInt) != 0 {
			mnfst.MetaData = tmpInt
		}

		// Grab the customer reference value
		if len(form.Value["customer_ref"]) != 0 && len(form.Value["customer_ref"][0]) != 0 {
			// Get the title value
			custref = form.Value["customer_ref"][0]
		}
	}

	// API origin request tasks
	if origin == config.Consts["api"] {
		// Grab the customer reference value
		if len(form.Value["customer_ref"]) != 0 && len(form.Value["customer_ref"][0]) != 0 {
			// Get the title value
			custref = form.Value["customer_ref"][0]
		}
	}

	// Get the event description
	event = "<none>"
	if len(form.Value["event"]) != 0 && len(form.Value["event"][0]) != 0 {
		// Get the title value
		event = form.Value["event"][0]
	}

	// Grab the set of files
	files := form.File["files"]

	// Iterate through the files
	for _, file := range files {
		// Set up a pipe that will pipe data from a writer to a reader
		//   writer -> duplicate of data written via io.MultiWriter sourced from an upload file
		//   reader -> read data sent to the cloud via a minio file transfer client
		pr, pw := io.Pipe()

		//  Set up a multiwriter 'mw' what will duplicate its inbound data from an upload file to:
		//    1. blk3hshr (hashing algorthm) and also...
		//    2. written to a pipe to consumed by a reader (that transfers that data to cloud storage)
		mw := io.MultiWriter(blk3hshr, pw)

		// Pick off the filename of the file path being processed
		filename := filepath.Base(file.Filename)

		// Construct the upload filename
		updFname := fmt.Sprintf("%v_%v", mnfst.RequestID, filename)

		// Open the file for reading
		rdFile, err := file.Open()
		if err != nil {
			// error opening an uploaded file
			log.Printf("ERROR: %v - error opening an uploaded file. See: %v\n",
				utils.FileLine(),
				err)

			errMap["msg"] = "error opening an uploaded file"
			c.JSON(http.StatusBadRequest, errMap)
			return
		}

		// Hash the request body data
		//   -> start a goroutine that will send data to be hashed (writer)
		go func() {
			_, err = io.Copy(mw, rdFile)
			if err != nil {
				// error hashing a file's contents
				log.Printf("ERROR: %v - error hashing a file's: %v contents. See: %v",
					utils.FileLine(),
					file.Filename,
					err)

				errMap["msg"] = "error hashing a file's contents"
				c.JSON(http.StatusBadRequest, errMap)
				return
			}

			pw.Close()
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
		defer cancel()

		_, err = hlr.SpacesClient.PutObjectWithContext(
			ctx,
			config.Consts["bucket_uploads"],
			updFname,
			pr,
			-1,
			minio.PutObjectOptions{ContentType: "application/octet-stream"})
		if err != nil {
			// error uploading a file to spaces
			log.Printf("ERROR: %v - error uploading a file to spaces. See: %v\n",
				utils.FileLine(),
				err)

			errMap["msg"] = "error saving a file"
			c.JSON(http.StatusInternalServerError, errMap)
			return
		}

		// Generate the hash of the uploaded file
		sum := blk3hshr.Sum(nil)
		hsh := fmt.Sprintf("%x", sum[:32])

		// Save filenames and hashes to the manifest
		tMap := make(map[string]string)
		tMap["hash"] = hsh
		tMap["filename"] = updFname
		mnfst.Contents = append(mnfst.Contents, tMap)

		// TODO: maybe update the meta data of the file just uploaded to Spaces with
		// TODO:   .... its hash? Just fire off a goroutine?

		// close the file being processed
		rdFile.Close()

		// close the pipe reader
		pr.Close()

		// Reset the hasher
		blk3hshr.Reset()
	}

	// Encode manifest as json
	_, mnbytes, err := utils.ToJSON(mnfst)
	if err != nil {
		// error encoding the manifest object as json bytes
		log.Printf("ERROR: %v - error encoding the manifest object as json bytes. See: %v\n",
			utils.FileLine(),
			err)

		errMap["msg"] = "error encoding the manifest object as json bytes"

		c.JSON(http.StatusInternalServerError, errMap)
		return
	}

	// Hash the json encoded document
	blk3hshr.Reset() // Rest the blake3 hasher
	_, err = blk3hshr.Write(mnbytes)
	if err != nil {
		// error hashing the manifest json object
		log.Printf("ERROR: %v - error hashing the manifest json object. See: %v\n",
			utils.FileLine(),
			err)
		errMap["msg"] = "error hashing the manifest json object"

		c.JSON(http.StatusInternalServerError, errMap)
		return
	}
	mnsum := blk3hshr.Sum(nil)

	// ********************************
	// Write the hash to the blockchain
	// Call a smart contract
	ctxSC, cnclSC := context.WithCancel(context.Background())
	defer cnclSC()

	// Generate an ABI object (for my deployed smart contract) using a file accessed via the web
	myABIFile := hlr.GoChainCntrABIURL
	abi, err := web3.GetABI(myABIFile)
	if err != nil {
		// error generating an ABI object
		log.Printf("ERROR: %v - error generating an ABI object. See: %v\n",
			utils.FileLine(),
			err)

		errMap["msg"] = "error generating an ABI object"

		c.JSON(http.StatusInternalServerError, errMap)
		return
	}

	// Invoke the previously deployed smart contract with some parameters
	txnSC, err := web3.CallTransactFunction(
		ctxSC,
		hlr.GoChainNetwork,
		*abi,
		config.Cfg.GoCntrtLogAddr,
		config.Cfg.GoChainPrivKey,
		"postObj",
		big.NewInt(0),
		fmt.Sprintf("%x", mnsum[:32]),
		mnfst.RequestID,
		custref,
	)
	if err != nil {
		// error calling the GoChain smart contract
		log.Printf("ERROR: %v - error calling the GoChain smart contract. See: %v\n",
			utils.FileLine(),
			err)

		errMap["msg"] = "error generating an ABI object"

		c.JSON(http.StatusInternalServerError, errMap)
		return
	}

	// Log blockwrite data to the database
	go hlr.logblockwrite(
		custid,
		origin,
		event,
		mnfst.RequestID,
		mnfst,
		fmt.Sprintf("0x%x", mnsum[:32]),
		fmt.Sprintf("0x%x", txnSC.Hash),
		fmtTxn(txnSC),
		utils.FileName(),
		&bwMtx)

	// Fetch transaction receipt data and update the blockwrite data referred to above
	go hlr.saveTxnReceipt(
		txnSC,
		mnfst.RequestID,
		&bwMtx)

	rspMap["msg"] = "hash written to the blockchain"
	rspMap["txnid"] = fmt.Sprintf("0x%x", txnSC.Hash)
	c.JSON(http.StatusOK, rspMap)

	// TODO: create a manifest json document including the filenames and their respective hashes

	// TODO:  --> create a tarball of each data file and the package manifest
	// TODO:  --> store that tarball in it's resting place (rpachain storage, IPFS, customer location)
}

// fmtTxn generates a map formatting the transaction object for human viewing
func fmtTxn(txn *web3.Transaction) map[string]string {
	var retMap = make(map[string]string)

	log.Printf("DEBUG: %v - dump out the web3 transaction\n", utils.FileLine())

	retMap["nonce"] = fmt.Sprintf("%d", txn.Nonce)
	retMap["gasprice"] = fmt.Sprintf("%d", txn.GasPrice)
	retMap["gaslimit"] = fmt.Sprintf("%d", txn.GasLimit)
	retMap["to"] = fmt.Sprintf("0x%x", txn.To)
	retMap["value"] = fmt.Sprintf("%d", txn.Value)
	retMap["input_string"] = fmt.Sprintf("%s", string(txn.Input))
	retMap["input_hex"] = fmt.Sprintf("0x%x", txn.Input)
	retMap["from"] = fmt.Sprintf("0x%x", txn.From)
	retMap["v"] = fmt.Sprintf("%d", txn.V)
	retMap["r"] = fmt.Sprintf("%d", txn.R)
	retMap["s"] = fmt.Sprintf("%d", txn.S)
	retMap["hash"] = fmt.Sprintf("%d", txn.V)
	retMap["blocknumber"] = fmt.Sprintf("%d", txn.BlockNumber)
	retMap["blockhash"] = fmt.Sprintf("0x%x", txn.BlockHash)
	retMap["transactionindex"] = fmt.Sprintf("%d", txn.TransactionIndex)

	return retMap
}
