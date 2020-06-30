package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/danoand/rpachain-api/config"
	mdl "github.com/danoand/rpachain-api/models"
	"github.com/danoand/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// TallyBlockWritesByDay returns the number of blockwrites by day
//    for the last seven days
func (hlr *HandlerEnv) TallyBlockWritesByDay(c *gin.Context) {
	var (
		err             error
		ok              bool
		now             time.Time
		then            time.Time
		thenStr, custID string
		blkWrts         []mdl.BlockWrite
		tallyMap        = make(map[string]int)
		rtArrDays       []string
		rtArrNum        []int
		rtMap           = make(map[string]interface{})
		rsp             = make(map[string]interface{})
	)

	// Get the customer document id
	custID, err = GetGinContextValStr(c.Copy(), config.Consts["cxtCustomerIDKey"])
	if err != nil {
		// error grabbing the customer document id
		log.Printf("ERROR: %v - error grabbing the customer document id from the gin context. See: %v\n",
			utils.FileLine(),
			err)
		rsp["msg"] = "an error occurred, please try again"
		c.JSON(http.StatusInternalServerError, rsp)
		return
	}

	// Determine the current time and 7 days ago
	now = time.Now()
	then = now.Add(-time.Hour * 24 * 7)

	thenStr = then.Format(time.RFC3339)[:10]

	// Fetch the objects from the database
	sel := bson.D{{"timestamp", bson.D{{"$gt", thenStr}}}, {"customerid", custID}}
	csr, err := hlr.CollBlockWrites.Find(context.TODO(), sel)
	if err != nil {
		// error fetching blockwrites from the database
		log.Printf("ERROR: %v - error fetching blockwrites from the database. See: %v\n",
			utils.FileLine(),
			err)
		rsp["msg"] = "an error occurred, please try again"
		c.JSON(http.StatusInternalServerError, rsp)
		return
	}

	// Grab all blockwrites from the query
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

	// Iterate the block writes and tally a count of objects
	for _, elm := range blkWrts {
		date1 := elm.TimeStamp[:10]
		_, ok = tallyMap[date1]
		if !ok {
			// new date - set counter to 1
			tallyMap[date1] = 1
			continue
		}

		tallyMap[date1] = tallyMap[date1] + 1
	}

	// Sort the date keys
	var dteArr1 []string
	// Iterate through the map and create an array of keys (dates)
	for k := range tallyMap {
		dteArr1 = append(dteArr1, k)
	}

	// Sort the array of dates
	sort.Strings(dteArr1)

	// Iterate through the sorted array of dates and construct return elements
	for _, elm := range dteArr1 {
		date2 := fmt.Sprintf("%v/%v", elm[5:7], elm[8:10])
		rtArrDays = append(rtArrDays, date2)       // houses days
		rtArrNum = append(rtArrNum, tallyMap[elm]) // houses # of objects stored that day
	}

	// Construct the return object
	rtMap["labels"] = rtArrDays
	rtMap["series"] = []string{"Block Updates"}
	rtMap["data"] = rtArrNum

	// Return to caller
	rsp["msg"] = "returning metrics"
	rsp["content"] = rtMap
	c.JSON(http.StatusOK, rsp)
}
