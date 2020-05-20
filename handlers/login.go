package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	mdl "github.com/danoand/rpachain-api/models"
	"github.com/globalsign/mgo/bson"

	ginsession "github.com/go-session/gin-session"

	"github.com/danoand/utils"
	"github.com/gin-gonic/gin"
)

// Login validates a login session
func (hlr *HandlerEnv) Login(c *gin.Context) {
	var err error
	var pAcct mdl.Account
	var acct mdl.Account
	var rsp = make(map[string]interface{})

	// Get parameters
	bbytes, err := c.GetRawData()
	if err != nil {
		// error grabbing request body data
		log.Printf("ERROR: %v - error grabbing request body data. See: %v\n",
			utils.FileLine(),
			err)

		rsp["msg"] = "error grabbing request data"
		c.JSON(http.StatusBadRequest, rsp)
		return
	}

	// Parse the request body
	err = utils.FromJSONBytes(bbytes, &pAcct)
	if err != nil {
		// error parsing the request body data
		log.Printf("ERROR: %v - error parsing the request body data. See: %v\n",
			utils.FileLine(),
			err)

		rsp["msg"] = "error parsing the request body data"
		c.JSON(http.StatusInternalServerError, rsp)
		return
	}

	// Validate the inbound data
	if len(pAcct.Username) == 0 || len(pAcct.Password) == 0 {
		// missing username or password
		log.Printf("ERROR: %v - missing username or password\n",
			utils.FileLine())

		rsp["msg"] = "missing username or password"
		c.JSON(http.StatusBadRequest, rsp)
		return
	}

	// Execute a round trip to the database
	ctx, cncl := context.WithTimeout(context.Background(), 5*time.Second)
	defer cncl()

	rslt := hlr.CollAccounts.FindOne(ctx, bson.M{"username": pAcct.Username})

	// Decode the fetch into a go object
	err = rslt.Decode(&acct)
	if err != nil {
		// error fetching a user account
		log.Printf("ERROR - %v - error fetching a user account. See: %v\n",
			utils.FileLine(),
			err)

		rsp["msg"] = "error fetching a user account"
		c.JSON(http.StatusBadRequest, rsp)
		return
	}

	// Does the password parameter match the account password?
	if pAcct.Password != acct.Password {
		// incorrect password
		log.Printf("ERROR - %v - incorrect password\n",
			utils.FileLine())

		rsp["msg"] = "incorrect password"
		c.JSON(http.StatusBadRequest, rsp)
		return
	}

	// Set session elements
	store := ginsession.FromContext(c)
	store.Set("docid", acct.ID.Hex())
	store.Set("username", acct.Username)

	rsp["msg"] = "you are now logged in"
	rsp["content"] = map[string]string{"docid": acct.ID.Hex(), "username": acct.Username}
	c.JSON(http.StatusOK, rsp)
}
