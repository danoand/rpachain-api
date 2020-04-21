package models

import (
	"github.com/danoand/gotomate-api/hash"
	"github.com/globalsign/mgo/bson"
	"github.com/gochain/web3"
)

// BlockWrite models the output of block write to the blockchain
type BlockWrite struct {
	ID               bson.ObjectId    `bson:"id" json:"docid"`
	Hash             string           `bson:"hash" json:"hash"`
	Manifest         hash.Manifest    `bson:"manifest" json:"manifest"`
	BlockTransaction web3.Transaction `bson:"transaction" json:"transaction"`
	BlockDataRef     string           `bson:"blockdataref" json:"blockdataref"`
}
