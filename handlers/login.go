package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	mdl "github.com/danoand/rpachain-api/models"

	"github.com/danoand/utils"
	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
)

// Login validates a login session
func (hlr *HandlerEnv) Login(c *gin.Context) {
	var err error
	var acct mdl.Account
	var eResp = make(map[string]interface{})

	log.Printf("DEBUG: %v - in Login handler\n", utils.FileLine())

	// Get parameters
	bbytes, err := c.GetRawData()
	if err != nil {
		// error grabbing request body data
		log.Printf("ERROR: %v - error grabbing request body data. See: %v\n",
			utils.FileLine(),
			err)

		eResp["msg"] = "error grabbing request data"
		c.JSON(http.StatusBadRequest, eResp)
		return
	}

	// Parse the request body
	err = utils.FromJSONBytes(bbytes, &acct)
	if err != nil {
		// error parsing the request body data
		log.Printf("ERROR: %v - error parsing the request body data. See: %v\n",
			utils.FileLine(),
			err)

		eResp["msg"] = "error parsing the request body data"
		c.JSON(http.StatusInternalServerError, eResp)
		return
	}

	// Validate the inbound data
	if len(acct.Username) == 0 || len(acct.Password) == 0 {
		// missing username or password
		log.Printf("ERROR: %v - missing username or password\n",
			utils.FileLine(),
			err)

		eResp["msg"] = "missing username or password"
		c.JSON(http.StatusBadRequest, eResp)
		return
	}

	// Execute a round trip to the database
	ctx, cncl := context.WithTimeout(context.Background(), 5*time.Second)
	defer cncl()

	rslt := hlr.CollAccounts.FindOne(ctx, bson.M{"username": acct.Username})

	// Decode the fetch into a go object
	err = rslt.Decode(&acct)
	if err != nil {
		// error fetching a user account
		log.Printf("ERROR - %v - error fetching a user account. See: %v\n",
			utils.FileLine(),
			err)

		eResp["msg"] = "error fetching a user account"
		c.JSON(http.StatusBadRequest, eResp)
		return
	}

	c.JSON(200, gin.H{
		"message": "you are now logged in",
	})
}
