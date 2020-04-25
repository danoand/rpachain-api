package config

import (
	"log"

	"github.com/danoand/utils"
	"github.com/kelseyhightower/envconfig"
)

// Specification houses configuration variables sourced from environment variables
type Specification struct {
	MGDBURLString       string `default:"mongodb+srv://dbdevuser:%v@cluster0-6uy29.mongodb.net/test?retryWrites=true&w=majority"`
	MGDBPassword        string `required:"true"`
	GoChainURL          string `default:"https://testnet-rpc.gochain.io/"`
	GoChainPrivKey      string `default:"0xcfa2b75c32a191e50a5612085dafac36c42e2ff6b46e110642e7ee45b916cc6b"`
	GoCntrtLogAddr      string `default:"0x30F7F8A09fAB59299588CceF5d410e99CeaAD9C8"`
	SpacesAccessKey     string `required:"true"`
	SpacesSecretKey     string `required:"true"`
	WrkIsWorkerInstance bool   `default:"true"` // indicates if this instance is a worker (as opposed to api instance)
}

// Cfg contains the environment variable information read from the execution environment
var Cfg Specification

// Read in environment variable information at initialization time
func init() {
	err := envconfig.Process("RPCH", &Cfg)
	if err != nil {
		log.Fatalf("FATAL: %v - error importing environment variables. See: %v\n", utils.FileLine(), err)
	}
}
