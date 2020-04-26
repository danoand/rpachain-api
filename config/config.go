package config

import (
	"log"

	"github.com/danoand/utils"
	"github.com/kelseyhightower/envconfig"
)

// Specification houses configuration variables sourced from environment variables
type Specification struct {
	MGDBURLString  string `default:"mongodb+srv://dbdevuser:%v@cluster0-6uy29.mongodb.net/test?retryWrites=true&w=majority"`
	MGDBPassword   string `required:"true"`
	GoChainURL     string `default:"https://testnet-rpc.gochain.io/"`
	GoChainPrivKey string `default:"0xcfa2b75c32a191e50a5612085dafac36c42e2ff6b46e110642e7ee45b916cc6b"`
	// Contract deployment txn id: 0xad90db5d1a37e64376ca0528fcc6a83cce5b1e9a3bba76455972188222e69214
	GoCntrtLogAddr      string `default:"0x584a428d76a8f2943F3B806fAC8c458F6107a789"`
	SpacesAccessKey     string `required:"true"`
	SpacesSecretKey     string `required:"true"`
	WrkIsWorkerInstance bool   `default:"false"` // indicates if this instance is a worker (as opposed to api instance)
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
