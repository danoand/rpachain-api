package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"

	config "github.com/danoand/rpachain-api/config"
	mdl "github.com/danoand/rpachain-api/models"
	"github.com/danoand/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	prmtv "go.mongodb.org/mongo-driver/bson/primitive"
)

// GetOneBlockWrite fetches a blockwrite from the database
func (hlr *HandlerEnv) GetOneBlockWrite(c *gin.Context) {
	var (
		err      error
		docid    string
		blkWrt   mdl.BlockWrite
		rsp      = make(map[string]interface{})
		rspMap   = make(map[string]interface{})
		arrMeta  []map[string]interface{}
		arrFiles []map[string]interface{}
	)

	// Grab the docid from the inbound request
	docid = c.Param("docid")
	// Generate an object id from the docid parameter
	_, err = prmtv.ObjectIDFromHex(docid)
	if err != nil {
		// error generating an objectid from the request
		log.Printf("ERROR: %v - error generating an objectid from the request. See: %v\n",
			utils.FileLine(),
			err)
		rsp["msg"] = "an error occurred, please try again"
		c.JSON(http.StatusInternalServerError, rsp)
		return
	}

	// Fetch a blockwrite
	err = hlr.CollBlockWrites.FindOne(context.TODO(), bson.D{{"requestid", docid}}).Decode(&blkWrt)
	if err != nil {
		// error fetching a blockwrite object
		log.Printf("ERROR: %v - error fetching a blockwrite object. See: %v\n",
			utils.FileLine(),
			err)
		rsp["msg"] = "error fetching a blockwrite object"
		c.JSON(http.StatusInternalServerError, rsp)
		return
	}

	// Construct the response object
	rspMap["source"] = "web" // TODO: source this data element
	rspMap["timestamp"] = blkWrt.TimeStamp
	rspMap["request_id"] = blkWrt.RequestID
	rspMap["config_nometadata"] = true
	rspMap["config_nofiles"] = true

	// Is there metadata to be displayed?
	if len(blkWrt.Manifest.MetaData) != 0 {
		// yes... add the metadata to the response
		rspMap["config_nometadata"] = false

		// Iterate through the metadata maps
		for key, val := range blkWrt.Manifest.MetaData {
			tMap := make(map[string]interface{})
			tMap["key"] = key
			tMap["value"] = val

			arrMeta = append(arrMeta, tMap)
		}

		// Add the array of meta data to the response object
		rspMap["metadata"] = arrMeta
	}

	// Are there files to be displayed
	if len(blkWrt.Manifest.Contents) != 0 {
		// yes... add the files to the response
		rspMap["config_nofiles"] = false

		// Iterate through the file maps
		for _, val := range blkWrt.Manifest.Contents {
			tMap := make(map[string]interface{})
			tMap["hash"] = val["hash"]
			tMap["filename_display"] = stripDocID(val["filename"])

			arrFiles = append(arrFiles, tMap)
		}

		// Add the array of files to the response object
		rspMap["files"] = arrFiles
	}

	// Grab Blockchain information
	var tMap = make(map[string]string)
	tMap["network"] = blkWrt.ChainNetwork
	tMap["block_number"] = blkWrt.BlockNumber
	tMap["block_url"] = fmt.Sprintf("%v%v/%v",
		config.Consts["gochain_testnet_explorer"],
		"block",
		blkWrt.BlockNumber)
	rspMap["block_info"] = tMap

	// Return the object to the client
	rsp["msg"] = "your blockwrite"
	rsp["content"] = rspMap
	c.JSON(http.StatusOK, rsp)
}
