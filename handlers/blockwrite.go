package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/danoand/utils"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
	"github.com/tidwall/pretty"
)

// GenNewKeccak256 generates a Keccak256 hash for a given byte slice
func GenNewKeccak256(b []byte) string {
	hash := crypto.Keccak256Hash(b)
	return hash.Hex()
}

// BlockWrite invoke a smart contract to write a hash to the EVM log and add data file to the IPFS
func (hlr *HandlerEnv) BlockWrite(c *gin.Context) {
	var err error
	var reqBytes, jBytes []byte
	var gotoMetaMap = make(map[string]interface{})
	var tmpMap = make(map[string]interface{})
	var errMap = make(map[string]string)

	gotoMetaMap["event"] = "blockwrite"
	gotoMetaMap["time"] = time.Now().Format(time.RFC3339)

	// Fetch the request body (assume JSON data)
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

	// Parse the JSON bytes into a go map
	err = utils.FromJSONBytes(reqBytes, &tmpMap)
	if err != nil {
		// error fetching the request body
		log.Printf("ERROR: %v - error fetching the request body. See: %v\n",
			utils.FileLine(),
			err)
		errMap["msg"] = "error accessing the inbound request data"

		c.JSON(http.StatusInternalServerError, errMap)
		return
	}

	// Encode the map into JSON
	_, jBytes, err = utils.ToJSON(tmpMap)
	if err != nil {
		// error encoding request body as json for hashing
		log.Printf("ERROR: %v - error encoding request body as json for hashing. See: %v\n",
			utils.FileLine(),
			err)
		errMap["msg"] = "error encoding request body as json for hashing"

		c.JSON(http.StatusInternalServerError, errMap)
		return
	}

	// Pretty up the encoded JSON
	jBytes = pretty.Pretty(jBytes)

	// Hash the passed JSON
	fileHash := GenNewKeccak256(jBytes)
	gotoMetaMap["filehash"] = fileHash

	// Hash the gotomate meta data map
	_, jBytes, err = utils.ToJSON(gotoMetaMap)
	if err != nil {
		// error encoding the gotomate meta data map
		log.Printf("ERROR: %v - error encoding the gotomate meta data map. See: %v\n",
			utils.FileLine(),
			err)
		errMap["msg"] = "error encoding the gotomate meta data map"

		c.JSON(http.StatusInternalServerError, errMap)
		return
	}
	mapHash := GenNewKeccak256(jBytes)


	// Create a string timestamp
	str := time.Now().In(hlr.TimeLocationCT).Format(time.RFC3339)

	// Execute a round trip to the database
	ctx, cncl := context.WithTimeout(context.Background(), 5*time.Second)
	defer cncl()

	// Execute update
	_, err = hlr.CollStatus.UpdateOne(
		ctx,
		bson.M{"type": "status"},
		bson.M{"$set": bson.M{"timestamp": str}})
	if err != nil {
		// error occurred writing during a db round trip transaction
		log.Printf("ERROR: %v - error occurred writing during a db round trip transaction. See: %v\n",
			utils.FileLine(),
			err)

		c.JSON(500, gin.H{
			"message": "error occurred accessing the database",
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "StatusOK",
	})
}
