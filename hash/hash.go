package hash

// Manifest models a description of data to memorialized
type Manifest struct {
	ID        string                   `bson:"id" json:"id"`
	TimeStamp string                   `bson:"timestamp" json:"timestamp"`
	MetaData  map[string]interface{}   `bson:"metadata" json:"metadata"`
	Contents  []map[string]interface{} `bson:"contents" json:"contents"`
}
