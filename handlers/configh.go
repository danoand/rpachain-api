package handlers

import (
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

// HandlerEnv houses config data needed for route handler execution
type HandlerEnv struct {
	TimeLocationCT *time.Location
	Client         *mongo.Client
	Database       *mongo.Database
	CollStatus     *mongo.Collection
}
