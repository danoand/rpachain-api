package middleware

import (
	"log"

	"github.com/danoand/rpachain-api/config"
	"github.com/danoand/utils"
	"github.com/gin-gonic/gin"
)

// APIAuth is middleware that authorizes a client API call and associates
//   a customer document id with the inbound request (in the the gin context)
//   NOTE: currently just a stub function standing in place of future code
func APIAuth() gin.HandlerFunc {
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
