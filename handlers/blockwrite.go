package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/danoand/gotomate-api/config"
	"github.com/danoand/gotomate-api/hash"
	"github.com/globalsign/mgo/bson"
	"github.com/gochain/web3"

	"github.com/danoand/utils"
	"github.com/gin-gonic/gin"
	"lukechampine.com/blake3"
)

// BlockWrite invoke a smart contract to write a hash to the EVM log and add data file to the IPFS
func (hlr *HandlerEnv) BlockWrite(c *gin.Context) {
	var (
		err      error
		mnfst    hash.Manifest
		reqBytes []byte
		errMap   = make(map[string]string)
		rspMap   = make(map[string]interface{})
	)

	// Fetch the request body
	reqBytes, err = c.GetRawData()
	if err != nil {
		// error fetching the request body
		log.Printf("ERROR: %v - error fetching the request body. See: %v\n",
			utils.FileLine(),
			err)
		errMap["msg"] = "error accessing the inbound request data"

		c.JSON(http.StatusInternalServerError, errMap)
		return
	}

	// Data missing?
	if len(reqBytes) == 0 {
		// missing request body data
		rspMap["msg"] = "missing data - no write action"
		c.JSON(http.StatusBadRequest, rspMap)
		return
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

	rspMap["msg"] = "hash written to the blockchain"
	rspMap["txnid"] = fmt.Sprintf("0x%x", txnSC.Hash)
	c.JSON(http.StatusOK, rspMap)
}
