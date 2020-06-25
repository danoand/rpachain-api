package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/danoand/rpachain-api/config"
	"github.com/danoand/utils"
	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
)

// APIAuth is middleware that authorizes a client API call and associates
//   a customer document id with the inbound request (in the the gin context)
//   NOTE: currently just a stub function standing in place of future code
func (hlr *HandlerEnv) APIAuth() gin.HandlerFunc {
	return func(c *gin.Context) {

		// Set example variable
		c.Set(config.Consts["cxtCustomerIDKey"], config.Consts["stub_custid"]) // string
		log.Printf("DEBUG: %v - setting a stub customer document id of: %v\n",
			utils.FileLine(),
			config.Consts["stub_custid"])

		// Transfer control to next handler
		c.Next()
	}
}

// WebAuth is middleware that authorizes a web request
func (hlr *HandlerEnv) WebAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		var docid, username string
		var result bson.M

		// Make a copy of the gin context
		cpy := c.Copy()

		// Get the request headers
		docid = cpy.GetHeader("X-docid")
		username = cpy.GetHeader("X-username")

		// Grab the account from the database
		docID, err := primitive.ObjectIDFromHex(docid)
		if err != nil {
			// error handling a document id
			log.Printf("ERROR: %v - error handling a document id. See: %v\n",
				utils.FileLine(),
				err)

			c.AbortWithStatusJSON(
				http.StatusInternalServerError,
				gin.H{"msg": "an error occurred; please log in again"})
			return
		}

		// Find the account document in the database
		ctx, cncl := context.WithTimeout(context.Background(), 10*time.Second)
		defer cncl()

		err = hlr.CollAccounts.FindOne(ctx, bson.M{"_id": docID}).Decode(&result)
		if err != nil {
			// error finding an account document
			log.Printf("ERROR: %v - error finding an account document. See: %v\n",
				utils.FileLine(),
				err)

			c.AbortWithStatusJSON(
				http.StatusInternalServerError,
				gin.H{"msg": "an error occurred; please log in again"})
			return
		}

		// Validate the account document
		if result["username"] != username || result["status"] != "active" {
			// invalid or inactive account
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{"msg": "invalid or inactive account; please log in again"})
			return
		}

		c.Next()
	}
}

// OriginWeb is middleware that sets the gin context indicating
//   that the request orginates from the web app
func (hlr *HandlerEnv) OriginWeb() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set the request origin to "web"
		c.Set(config.Consts["cxtRequestOrigin"], config.Consts["web"])

		c.Next()
	}
}

// OriginAPI is middleware that sets the gin context indicating
//   that the request orginates from a call to the API
func (hlr *HandlerEnv) OriginAPI() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set the request origin to "web"
		c.Set(config.Consts["cxtRequestOrigin"], config.Consts["api"])

		c.Next()
	}
}
