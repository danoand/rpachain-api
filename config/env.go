package config

import (
	"log"

	"github.com/danoand/utils"
	"github.com/kelseyhightower/envconfig"
)

// Specification houses configuration variables sourced from environment variables
type Specification struct {
	MGDBURLString string `default:"mongodb+srv://dbdevuser:%v@cluster0-6uy29.mongodb.net/test?retryWrites=true&w=majority"`
	MGDBPassword  string `required:"true"`
}

// Cfg contains the environment variable information read from the execution environment
var Cfg Specification

// Read in environment variable information at initialization time
func init() {
	err := envconfig.Process("GOTO", &Cfg)
	if err != nil {
		log.Fatalf("FATAL: %v - error importing environment variables. See: %v\n", utils.FileLine(), err)
	}
}
