package handlers

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/danoand/rpachain-api/config"
	mdl "github.com/danoand/rpachain-api/models"
	"github.com/globalsign/mgo/bson"
	"github.com/minio/minio-go"

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
	var goErr error
	var ok bool
	var req mdl.BlockWrite
	var tw *tar.Writer

	help := wrk.HelperFor(ctx)

	// Grab the job id
	jid := help.Jid()

	log.Printf("WRKR: %v - started working on job %s\n",
		utils.FileLine(),
		help.Jid())

	defer func() {
		log.Printf("WRKR: %v - ending work on job %s with possible errors: [%v, %v]\n",
			utils.FileLine(),
			help.Jid(),
			err,
			goErr)
	}()

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
	err = rslt.Decode(&req)
	if err != nil {
		// error decoding the database document into an object
		log.Printf("WRKR: %v - ERROR - %v - error decoding the database document into an object. See: %v\n",
			jid,
			utils.FileLine(),
			docid)

		return fmt.Errorf("error decoding the database document into an object: %v", err)
	}

	// Iterate through the files to be tar'ed
	fnames := []string{}
	for _, fMap := range req.Manifest.Contents {
		fnames = append(fnames, fMap["filename"])
	}

	// Declare an io.Pipe that will transfer data from an io.Writer to an io.Reader ("piping" it)
	pr, pw := io.Pipe()

	// Define a tar writer
	tw = tar.NewWriter(pw)

	// Iterate through the spaces files and send them through the tar "writer"
	go func() {
		// Close the pipe and tar writers at the end of the goroutine
		defer pw.Close()
		defer tw.Close()

		for _, fname := range fnames {

			// Create a context for the bucket read
			ctx1, cancel1 := context.WithTimeout(context.Background(), 240*time.Second)
			defer cancel1()

			// Cue up the bucket object for reading
			object, err := hlr.SpacesClient.GetObjectWithContext(
				ctx1,
				config.Consts["bucket_uploads"],
				fname,
				minio.GetObjectOptions{})
			if err != nil {
				// error occurred reading a file from the remote bucket
				log.Printf("WRKR: %v - ERROR - %v - error occurred reading a file: %v from the remote bucket. See: %v\n",
					jid,
					utils.FileLine(),
					fname,
					err)

				goErr = fmt.Errorf("error occurred reading a file: %v from the remote bucket: %v", fname, err)
			}

			// Copy the file to the tar writer for processing
			_, err = io.Copy(tw, object)
			if err != nil {
				// error occurred tar'ing a file
				log.Printf("WRKR: %v - ERROR - %v - error occurred tar'ing a file: %v read from the remote bucket. See: %v\n",
					jid,
					utils.FileLine(),
					fname,
					err)

				goErr = fmt.Errorf("error occurred tar'ing a file: %v from the remote bucket: %v", fname, err)
			}
		}
	}()

	// Define a context for the remote bucket put operation
	ctx2, cancel2 := context.WithTimeout(context.Background(), 240*time.Second)
	defer cancel2()

	// Write the tar ball file to the remote bucket
	n, err := hlr.SpacesClient.PutObjectWithContext(
		ctx2,
		config.Consts["bucket_store"],
		fmt.Sprintf("%v.tar", docid),
		pr,
		-1,
		minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		// error putting the tar file object to the remote bucket
		log.Printf("WRKR: %v - ERROR - %v - error putting the tar file object: %v to the remote bucket. See: %v\n",
			jid,
			utils.FileLine(),
			fmt.Sprintf("%v.tar", docid),
			err)

		return fmt.Errorf("error putting the tar file object: %v to the remote bucket: %v",
			fmt.Sprintf("%v.tar", docid),
			err)
	}

	// Was there an error during the tar'ing process that was not picked up?
	if goErr != nil {
		// error occurred during the tar'ing process
		log.Printf("WRKR: %v - ERROR - %v - error occurred during the tar'ing process. See: %v\n",
			jid,
			utils.FileLine(),
			goErr)

		return fmt.Errorf("error occurred during the tar'ing process: %v", goErr)
	}

	log.Printf("WRKR: %v - INFO - %v bytes of tar file %v written to the remote bucket\n",
		utils.FileLine(),
		n,
		fmt.Sprintf("%v.tar", docid))

	return nil
}
