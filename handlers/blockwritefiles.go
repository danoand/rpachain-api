package handlers

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/danoand/rpachain-api/config"
	"github.com/globalsign/mgo/bson"
	"github.com/minio/minio-go"
	"lukechampine.com/blake3"

	"github.com/danoand/utils"
	"github.com/gin-gonic/gin"
)

// BlockWriteFiles hashes request files and writes the manifest hash to the blockchain
func (hlr *HandlerEnv) BlockWriteFiles(c *gin.Context) {
	var (
		err error
		// mnfst    hash.Manifest
		// reqBytes []byte
		errMap = make(map[string]string)
		// rspMap = make(map[string]interface{})
	)

	// Create a identifier for this request
	reqID := bson.NewObjectId().Hex()

	// Set up a Blake3 "hasher"
	blk3hshr := blake3.New(256, nil)

	// Process a multipart form style request
	// custRef := c.PostForm("ref") // sample: capture a reference number

	// Parse the multi-part request
	form, err := c.MultipartForm()
	if err != nil {
		// error parsing the request
		log.Printf("ERROR: %v - error parsing the request. See: %v\n",
			utils.FileLine(),
			err)

		errMap["msg"] = "error parsing the request"
		c.JSON(http.StatusBadRequest, errMap)
		return
	}

	// Grab the set of files
	files := form.File["files"]

	// Iterate through the files
	for _, file := range files {
		// Set up a pipe that will pipe data from a writer to a reader
		//   writer -> duplicate of data written via io.MultiWriter sourced from an upload file
		//   reader -> read data sent to the cloud via a minio file transfer client
		pr, pw := io.Pipe()

		//  Set up a multiwriter 'mw' what will duplicate its inbound data from an upload file to:
		//    1. blk3hshr (hashing algorthm) and also...
		//    2. written to a pipe to consumed by a reader (that transfers that data to cloud storage)
		mw := io.MultiWriter(blk3hshr, pw)

		// Pick off the filename of the file path being processed
		filename := filepath.Base(file.Filename)

		// Construct the upload filename
		updFname := fmt.Sprintf("%v_%v", reqID, filename)

		// Open the file for reading
		rdFile, err := file.Open()
		if err != nil {
			// error opening an uploaded file
			log.Printf("ERROR: %v - error opening an uploaded file. See: %v\n",
				utils.FileLine(),
				err)

			errMap["msg"] = "error opening an uploaded file"
			c.JSON(http.StatusBadRequest, errMap)
			return
		}

		// Hash the request body data
		//   -> start a goroutine that will send data to be hashed (writer)
		go func() {
			_, err = io.Copy(mw, rdFile)
			if err != nil {
				// error hashing a file's contents
				log.Printf("ERROR: %v - error hashing a file's: %v contents. See: %v",
					utils.FileLine(),
					file.Filename,
					err)

				errMap["msg"] = "error hashing a file's contents"
				c.JSON(http.StatusBadRequest, errMap)
				return
			}

			pw.Close()
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
		defer cancel()

		_, err = hlr.SpacesClient.PutObjectWithContext(
			ctx,
			config.Consts["bucket_uploads"],
			updFname,
			pr,
			-1,
			minio.PutObjectOptions{ContentType: "application/octet-stream"})
		if err != nil {
			// error uploading a file to spaces
			log.Printf("ERROR: %v - error uploading a file to spaces. See: %v\n",
				utils.FileLine(),
				err)

			errMap["msg"] = "error saving a file"
			c.JSON(http.StatusInternalServerError, errMap)
			return
		}

		sum := blk3hshr.Sum(nil)
		hsh := fmt.Sprintf("%x", sum[:32])

		log.Printf("DEBUG: %v - file: %v's hash is: %v\n",
			utils.FileLine(),
			updFname,
			hsh)

		// TODO: maybe update the meta data of the file just uploaded to Spaces with
		// TODO:   .... its hash? Just fire off a goroutine?

		// close the file being processed
		err = rdFile.Close()
		log.Printf("DEBUG: rdFile.Close() error: %v\n", err)

		// close the pipe reader
		err = pr.Close()
		log.Printf("DEBUG: r.Close() error: %v\n", err)

		// Reset the hasher
		blk3hshr.Reset()
	}

	c.JSON(http.StatusOK, "done uploading")

	// TODO: create a manifest json document including the filenames and their respective hashes

	// TODO: now that we have:
	// TODO:  1. uploaded files stored in the cloud (DO space)
	// TODO:  2. a Blake3 hash for each file
	// TODO:  --> create a tarball of each data file and the package manifest
	// TODO:  --> store that tarball in it's resting place (rpachain storage, IPFS, customer location)
}
