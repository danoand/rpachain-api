package models

// Manifest models a description of data to memorialized
type Manifest struct {
	RequestID         string                   `json:"id"`                // request id
	TimeStamp         string                   `json:"timestamp"`         // timestamp
	MetaData          map[string]interface{}   `json:"metadata"`          // general meta data
	Contents          []map[string]interface{} `json:"contents"`          // filenames and associated hashes
	CustomerReference map[string]interface{}   `json:"customerreference"` // general customer reference
}
