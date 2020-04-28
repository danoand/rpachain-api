package config

import (
	"log"
	"regexp"

	"github.com/danoand/utils"
	"github.com/kelseyhightower/envconfig"
)

// Specification houses configuration variables sourced from environment variables
type Specification struct {
	IsHerokuEnv    bool   `default:"false"`
	MGDBURLString  string `default:"mongodb+srv://dbdevuser:%v@cluster0-6uy29.mongodb.net/test?retryWrites=true&w=majority"`
	MGDBPassword   string `required:"true"`
	GoChainURL     string `default:"https://testnet-rpc.gochain.io/"`
	GoChainPrivKey string `default:"0xcfa2b75c32a191e50a5612085dafac36c42e2ff6b46e110642e7ee45b916cc6b"`
	// Contract deployment txn id: 0x4164faa2d1517ddb7044942a6c9243390a8f2951779844642dec355f55873f08
	GoCntrtLogAddr      string `default:"0x9322e08aCa890ED9fBf2Ba96Af87B9cF3187ab9F"`
	GoCntrtABIURL       string `default:"https://api.cacher.io/raw/b111c7d6e40a12daefbb/293719cb85094bc72614/BlockWriteSample.abi"`
	SpacesAccessKey     string `required:"true"`
	SpacesSecretKey     string `required:"true"`
	WrkIsWorkerInstance bool   `default:"false"` // indicates if this instance is a worker (as opposed to api instance)
}

// Cfg contains the environment variable information read from the execution environment
var Cfg Specification

// RgxFnamePrefix matches the prefix given to uploaded hashed files
var RgxFnamePrefix *regexp.Regexp

// Read in environment variable information at initialization time
func init() {
	err := envconfig.Process("RPCH", &Cfg)
	if err != nil {
		log.Fatalf("FATAL: %v - error importing environment variables. See: %v\n", utils.FileLine(), err)
	}

	RgxFnamePrefix = regexp.MustCompile(`^[a-f0-9]+_`)
}
