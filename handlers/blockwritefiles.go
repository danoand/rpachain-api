package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/danoand/rpachain-api/config"
	"github.com/danoand/rpachain-api/hash"
	"github.com/globalsign/mgo/bson"
	"github.com/gochain/web3"

	"github.com/danoand/utils"
	"github.com/gin-gonic/gin"
	"lukechampine.com/blake3"
)

// BlockWriteFiles hashes request files and writes the manifest hash to the blockchain
func (hlr *HandlerEnv) BlockWriteFiles(c *gin.Context) {
	var (
		err      error
		mnfst    hash.Manifest
		reqBytes []byte
		errMap   = make(map[string]string)
		rspMap   = make(map[string]interface{})
	)

	// Process a multipart form style request
	custRef := c.PostForm("ref") // sample: capture a reference number

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

	// Grab the set of files
	files := form.File["files"]

	// Iterate through the files
	for _, file := range files {
		// Pick off the filename of the file being processed
		filename := filepath.Base(file.Filename)
		minio.Put

	}

	// Set up a Blake3 "hasher"
	blk3hshr := blake3.New(256, nil)

	// Hash the request body data
	_, err = blk3hshr.Write(reqBytes)
	if err != nil {
		// error occurred hashing the request data
		log.Printf("ERROR: %v - error occurred hashing the request data. See: %v\n",
			utils.FileLine(),
			err)
		errMap["msg"] = "error occurred processing the request data"

		c.JSON(http.StatusInternalServerError, errMap)
		return
	}

	// Generate the hash value
	hsh := blk3hshr.Sum(nil)
	// Grab the hex representation of the first 32 bytes
	hshStr := fmt.Sprintf("%x", hsh[:32])

	// Update the hash manifest
	mnfst.ID = bson.NewObjectId().Hex()
	mnfst.TimeStamp = time.Now().In(hlr.TimeLocationCT).Format(time.RFC3339)
	tMap := make(map[string]interface{})
	tMap[hshStr] = "see file" // TODO: change this value?
	mnfst.Contents = append(mnfst.Contents, tMap)

	// Encode manifest as bson
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
		mnfst.ID,
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
	go hlr.logblockwrite(fmt.Sprintf("0x%x", txnSC.Hash), mnfst, txnSC, "")

	rspMap["msg"] = "hash written to the blockchain"
	rspMap["txnid"] = fmt.Sprintf("0x%x", txnSC.Hash)
	c.JSON(http.StatusOK, rspMap)
}
