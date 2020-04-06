package handlers

import (
	"context"
	"log"
	"time"

	"github.com/danoand/utils"
	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
)

// Status page: return ok if the web server is up
func (hlr *HandlerEnv) Status(c *gin.Context) {
	var ctx context.Context
	var err error

	// Create a string timestamp
	str := time.Now().In(hlr.TimeLocationCT).Format(time.RFC3339)

	// Execute a round trip to the database
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)
	_, err = hlr.CollStatus.UpdateOne(ctx, bson.M{"type": "status"}, bson.M{"timestamp": str})
	if err != nil {
		// error occurred writing during a db round trip transaction
		log.Printf("ERROR: %v - error occurred writing during a db round trip transaction. See: %v\n",
			utils.FileLine(),
			err)

		c.JSON(500, gin.H{
			"message": "error occurred accessing the database"
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "StatusOK",
	})
}
