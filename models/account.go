package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// Account models an application account
type Account struct {
	ID          primitive.ObjectID `bson:"_id" json:"id,omitempty"` // account id
	Username    string             `json:"username"`                // timestamp
	Password    string             `json:"password"`                // general meta data
	AccountName string             `json:"accountname"`             // account company name
}
