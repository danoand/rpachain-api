package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"

	mdl "github.com/danoand/rpachain-api/models"
	"github.com/danoand/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gin-gonic/gin"
)

// GetBlockWrites gets blockwrites for display on the dashboard
func (hlr *HandlerEnv) GetBlockWrites(c *gin.Context) {
	var err error
	var blkWrts []mdl.BlockWrite
	var rsp = make(map[string]interface{})

	// Fetch the block writes from the database
	opts := options.Find().SetSort(bson.D{{"manifest.timestamp", -1}})
	csr, err := hlr.CollBlockWrites.Find(context.TODO(), bson.D{}, opts)
	if err != nil {
		// error fetching blockwrites from the database
		log.Printf("ERROR: %v - error fetching blockwrites from the database. See: %v\n",
			utils.FileLine(),
			err)
		rsp["msg"] = "an error occurred, please try again"
		c.JSON(http.StatusInternalServerError, rsp)
		return
	}

	// Grab all blockwrites from teh query
	err = csr.All(context.TODO(), &blkWrts)
	if err != nil {
		// error accessing blockwrite data
		log.Printf("ERROR: %v - error accessing blockwrite data. See: %v\n",
			utils.FileLine(),
			err)
		rsp["msg"] = "an error occurred, please try again"
		c.JSON(http.StatusInternalServerError, rsp)
		return
	}

	// Iterate through the query results and construct a response
	var tmpCntArr []map[string]interface{}
	for _, elm := range blkWrts {
		var tmpCnt = make(map[string]interface{})

		// Construct a response array element
		tmpCnt["network"] = elm.ChainNetwork
		tmpCnt["timestamp"] = fmt.Sprintf("%v CST", elm.Manifest.TimeStamp[:19])
		tmpCnt["block"] = elm.BlockNumber
		tmpCnt["action"] = "Notarize content to the blockchain"
		tmpCnt["explorer_link"] = fmt.Sprintf("%vblock/%v",
			elm.ChainNetwork,
			elm.BlockNumber)

		// Add element to the response
		tmpCntArr = append(tmpCntArr, tmpCnt)
	}

	// Construct the final response
	rsp["msg"] = "blockwrites"
	rsp["content"] = tmpCntArr

	c.JSON(http.StatusOK, rsp)
}
