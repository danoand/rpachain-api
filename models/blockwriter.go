package models

// BlockWrite models the output of block write to the blockchain
type BlockWrite struct {
	ID               string            `json:"id"`               // unique id
	CustomerID       string            `json:"customerid"`       // customer id
	ChainNetwork     string            `json:"chainnetwork"`     // blockchain network updated
	RequestID        string            `json:"requestid"`        // api request id
	TimeStamp        string            `json:"timestamp"`        // blockchain tx time
	ContractAddress  string            `json:"contractaddress"`  // blockchain contract address
	TransactionHash  string            `json:"transactionhash"`  // blockchain transaction hash
	ManifestHash     string            `json:"manifesthash"`     // overall hash (of manifest document) written to block log
	Manifest         Manifest          `json:"manifest"`         // txn manifest
	BlockTransaction map[string]string `json:"blocktransaction"` // block txn execution data
	Function         string            `json:"function"`         // name of the go function executing the block txn
}
