package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	mdl "github.com/danoand/rpachain-api/models"
	"github.com/globalsign/mgo/bson"

	wrkclnt "github.com/contribsys/faktory/client"
	wrk "github.com/contribsys/faktory_worker_go"
	"github.com/danoand/utils"
	"github.com/gin-gonic/gin"
)

// FaktoryStatus returns the status of the Faktory client
func (hlr *HandlerEnv) FaktoryStatus(c *gin.Context) {
	var err error
	var rMap = make(map[string]interface{})

	// Create a Faktory client
	fakclnt, err := wrkclnt.Open()
	if err != nil {
		// error opening a Faktory client
		log.Printf("ERROR: %v - error opening a Faktory client. See: %v\n",
			utils.FileLine(),
			err)

		c.String(
			http.StatusInternalServerError,
			fmt.Sprintf(
				"ERROR: %v - error opening a Faktory client. See: %v\n",
				utils.FileLine(),
				err))

		return
	}

	rMap, err = fakclnt.Info()
	rMap["request_error"] = err
	rMap["msg"] = "response from the Faktory client"
	c.JSON(http.StatusOK, rMap)
}

// TestFunktion represents a simple job queued to a Faktory instance
func (hlr *HandlerEnv) TestFunktion(ctx context.Context, args ...interface{}) error {
	help := wrk.HelperFor(ctx)

	log.Printf("WRKR: %v - started working on on job %s\n",
		utils.FileLine(),
		help.Jid())

	log.Printf("WRKR: %v - ending work on job %s\n",
		utils.FileLine(),
		help.Jid())

	return nil
}

// TarRequest creates and stores a tar ball of a request's hashed files
func (hlr *HandlerEnv) TarRequest(ctx context.Context, args ...interface{}) error {
	var err error
	var ok bool
	var req mdl.BlockWrite

	help := wrk.HelperFor(ctx)

	// Grab the job id
	jid := help.Jid()

	log.Printf("WRKR: %v - started working on on job %s\n",
		utils.FileLine(),
		help.Jid())

	// Get the request document id
	if len(args) == 0 {
		// missing the request document id job argument
		log.Printf("WRKR: %v - ERROR - %v - missing the request document id job argument\n",
			jid,
			utils.FileLine())

		return fmt.Errorf("missing the request document id job argument")
	}

	// Grab the request document id job argument
	dcid := args[0]

	// Assert the parameter type
	docid, ok := dcid.(string)
	if !ok {
		// invalid parameter type - expecting a string document id
		log.Printf("WRKR: %v - ERROR - %v - invalid parameter type - expecting a string document id\n",
			jid,
			utils.FileLine())

		return fmt.Errorf("invalid parameter type - expecting a string document id")
	}

	// Valid parameter?
	if !bson.IsObjectIdHex(docid) {
		// invalid document id
		log.Printf("WRKR: %v - ERROR - %v - invalid document id: %v\n",
			jid,
			utils.FileLine(),
			docid)

		return fmt.Errorf("invalid parameter type - expecting a string document id")

	}

	// Fetch the job request data
	ctx, cncl := context.WithTimeout(context.Background(), 5*time.Second)
	defer cncl()
	rslt := hlr.CollBlockWrites.FindOne(ctx, bson.M{"requestid": docid})
	if err != nil {
		// error fetching a request blockwrite object
		log.Printf("WRKR: %v - ERROR - %v - error fetching a request blockwrite object. See: %v\n",
			jid,
			utils.FileLine(),
			docid)

		return fmt.Errorf("error fetching a request blockwrite object: %v", err)
	}

	// Decode the fetch into a go object
	err = rslt.Decode(req)
	if err != nil {
		// error decoding the database document into an object
		log.Printf("WRKR: %v - ERROR - %v - error decoding the database document into an object. See: %v\n",
			jid,
			utils.FileLine(),
			docid)

		return fmt.Errorf("error decoding the database document into an object: %v", err)
	}

	return err
}
