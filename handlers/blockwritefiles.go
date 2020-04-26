package handlers

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/davecgh/go-spew/spew"

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
		err    error
		ok     bool
		dcVal  interface{}
		custid string
		mnfst  models.Manifest
		// reqBytes []byte
		errMap = make(map[string]string)
		rspMap = make(map[string]interface{})
	)

	mnfst.MetaData = make(map[string]interface{})

	// Get the customer document id from the gin context
	dcVal, ok = c.Get(config.Consts["cxtCustomerIDKey"])
	if !ok {
		// missing customer document id
		log.Printf("ERROR: %v - missing customer document id\n", utils.FileLine())

		errMap["msg"] = "an error occurred"
		c.JSON(http.StatusBadRequest, errMap)
	}
	custid, ok = dcVal.(string)
	if !ok {
		// unexpected parameter type - expecting a string
		log.Printf("ERROR: %v - unexpected parameter type - expecting a string: %v\n",
			utils.FileLine(),
			dcVal)

		errMap["msg"] = "an error occurred"
		c.JSON(http.StatusBadRequest, errMap)
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
		mnfst.Contents = append(
			mnfst.Contents,
			map[string]interface{}{
				"hash":     hsh,
				"filename": filename,
			})

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
	myABIFile := "https://dsfiles.dananderson.dev/files/LogObject.abi" // TODO: don't hardcode this
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
		0,
		"<cust_id>",
		mnfst.RequestID,
		fmt.Sprintf("%x", mnsum[:32]),
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
		mnfst.RequestID,
		mnfst,
		fmt.Sprintf("0x%x", mnsum[:32]),
		fmt.Sprintf("0x%x", txnSC.Hash),
		fmtTxn(txnSC),
		utils.FileName())

	rspMap["msg"] = "hash written to the blockchain"
	rspMap["txnid"] = fmt.Sprintf("0x%x", txnSC.Hash)
	c.JSON(http.StatusOK, rspMap)

	// TODO: create a manifest json document including the filenames and their respective hashes

	// TODO: now that we have:
	// TODO:  1. uploaded files stored in the cloud (DO space)
	// TODO:  2. a Blake3 hash for each file
	// TODO:  --> create a tarball of each data file and the package manifest
	// TODO:  --> store that tarball in it's resting place (rpachain storage, IPFS, customer location)
}

// fmtTxn generates a map formatting the transaction object for human viewing
func fmtTxn(txn *web3.Transaction) map[string]string {
	var retMap = make(map[string]string)

	log.Printf("DEBUG: %v - dump out the web3 transaction\n", utils.FileLine())
	spew.Dump(txn)

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
