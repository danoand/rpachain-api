package config

// Consts defines string constants that can be used in the app
var Consts = map[string]string{
	"bucket_uploads":           "rpachain-io-uploads",                  // spaces bucket for file uploads
	"bucket_store":             "test-rpachain-io",                     // spaces bucket to persist zip files // TODO: change this at some point
	"timezone":                 "America/Chicago",                      //
	"cxtCustomerIDKey":         "cxtCustomerID",                        // gin context key representing the customer document id
	"cxtRequestOrigin":         "cxtRequestOrigin",                     // Indicates where a request originates from (web, api)
	"web":                      "web",                                  //
	"api":                      "api",                                  //
	"stub_custid":              "5ea46028fc713915da7cf68d",             // temporary stub data (customer id)
	"localport_api":            "localhost:8080",                       // localhost port for the api instance
	"localport_wrk":            "localhost:8081",                       // localhost port for the worker instance
	"gochain_testnet_explorer": "https://testnet-explorer.gochain.io/", // testnet explorer url
}
