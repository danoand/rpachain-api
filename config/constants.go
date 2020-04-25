package config

// Consts defines string constants that can be used in the app
var Consts = map[string]string{
	"bucket_uploads":   "rpachain-io-uploads",      // spaces bucket for file uploads
	"timezone":         "America/Chicago",          //
	"cxtCustomerIDKey": "cxtCustomerID",            // gin context key representing the customer document id
	"stub_custid":      "5ea46028fc713915da7cf68d", // temporary stub data (customer id)
}
