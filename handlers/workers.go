package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"

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
