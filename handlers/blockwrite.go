package handlers

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/danoand/rpachain-api/config"
	"github.com/danoand/rpachain-api/models"
	"github.com/globalsign/mgo/bson"
	"github.com/gochain/web3"
	"github.com/minio/minio-go"

	"github.com/danoand/utils"
	"github.com/gin-gonic/gin"
	"lukechampine.com/blake3"
)

// BlockWrite invoke a smart contract to write a hash to the EVM log and add data file to the IPFS
func (hlr *HandlerEnv) BlockWrite(c *gin.Context) {
	var (
		err                              error
		custid, custref                  string
		origin                           string
		mnfst                            models.Manifest
		reqBytes                         []byte
		reqStr, reqRef                   string
		hshReq, hshRef, hshMfst          []byte
		hshReqStr, hshRefStr, hshMfstStr string
		tmpInt                           = make(map[string]interface{})
		tmpMap                           = make(map[string]string)
		errMap                           = make(map[string]string)
		rspMap                           = make(map[string]interface{})
		bwMtx                            sync.Mutex
	)

	custref = "N/A"

	// Get the customer document id from the gin context
	custid, err = GetGinContextValStr(c.Copy(), config.Consts["cxtCustomerIDKey"])
	if err != nil {
		// error grabbing the customer id from the gin context
		log.Printf("ERROR: %v - error grabbing the customer id from the gin context. See: %v\n",
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

	// Is the request body empty?
	if len(reqBytes) == 0 {
		// missing request body data
		rspMap["msg"] = "missing data - no write action"
		c.JSON(http.StatusBadRequest, rspMap)
		return
	}

	// Parse the request body as json
	err = utils.FromJSONBytes(reqBytes, &tmpMap)
	if err != nil {
		// error parsing the json body
		log.Printf("ERROR: %v - error parsing the json body. See: %v\n",
			utils.FileLine(),
			err)
		errMap["msg"] = "error parsing the json body"

		c.JSON(http.StatusInternalServerError, errMap)
		return
	}

	// Save content as bytes to be hashed below
	reqStr = tmpMap["content"]
	reqRef = tmpMap["customer_reference"]

	// Web origin request data elements
	if origin == config.Consts["web"] {
		// Assign the inbound data to the manifest
		tmpInt["event"] = tmpMap["event"]
		tmpInt["meta_data_01"] = tmpMap["meta_data_01"]
		tmpInt["content"] = tmpMap["content"]
		mnfst.MetaData = tmpInt

		// Store a customer reference value if provided
		if len(tmpMap["customer_ref"]) != 0 {
			reqRef = tmpMap["customer_ref"]
			mnfst.CustomerReference = make(map[string]interface{})
			mnfst.CustomerReference["web_customer_ref"] = reqRef
		}
	}

	// Set up a Blake3 "hasher" object
	blk3hshr := blake3.New(256, nil)

	//* Hash the request content
	_, err = blk3hshr.Write([]byte(reqStr))
	if err != nil {
		// error occurred hashing the request data
		log.Printf("ERROR: %v - error occurred hashing the request data. See: %v\n",
			utils.FileLine(),
			err)
		errMap["msg"] = "error occurred processing the request data"

		c.JSON(http.StatusInternalServerError, errMap)
		return
	}
	// Generate the hash value of the request body
	hshReq = blk3hshr.Sum(nil)
	// Grab the hex representation of the first 32 bytes
	hshReqStr = fmt.Sprintf("%x", hshReq[:32])
	// Reset the hashing object (to hash another chunk of content)
	blk3hshr.Reset()

	//* Hash the customer reference value
	if len(reqRef) != 0 {
		_, err = blk3hshr.Write([]byte(reqRef))
		if err != nil {
			// error occurred hashing the customer reference value
			log.Printf("ERROR: %v - error occurred hashing customer reference value. See: %v\n",
				utils.FileLine(),
				err)
			errMap["msg"] = "error occurred processing customer reference value"

			c.JSON(http.StatusInternalServerError, errMap)
			return
		}
		// Grab the hex representation of the first 32 bytes
		hshReqStr = fmt.Sprintf("%x", hshReq[:32])
		// Reset the hashing object (to hash another chunk of content)
		blk3hshr.Reset()
	}

	// Update the hash manifest
	mnfst.RequestID = bson.NewObjectId().Hex()

	// Create a context
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	// Write the request body data to a file in the storage bucket
	_, err = hlr.SpacesClient.PutObjectWithContext(
		ctx,
		config.Consts["bucket_uploads"],
		fmt.Sprintf("%v_request_body.dat", mnfst.RequestID),
		rdr,
		int64(rdr.Len()),
		minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		// error uploading a file to spaces
		log.Printf("ERROR: %v - error uploading a file to spaces for request: %v. See: %v\n",
			utils.FileLine(),
			mnfst.RequestID,
			err)

		errMap["msg"] = "error saving a file"
		c.JSON(http.StatusInternalServerError, errMap)
		return
	}

	// Update the hash manifest
	mnfst.TimeStamp = time.Now().In(hlr.TimeLocationCT).Format(time.RFC3339)
	// Save filenames and hashes to the manifest
	tMap := make(map[string]string)
	tMap["hash"] = hshStr
	tMap["filename"] = fmt.Sprintf("%v_request_body.dat", mnfst.RequestID)
	mnfst.Contents = append(mnfst.Contents, tMap)

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
		tmpMap["event"],
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
}
