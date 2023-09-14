package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

/*
Model for documents populating the monitor collection in MongoDB
*/
type MonitorEntry struct {
	ID        primitive.ObjectID `bson:"_id"`
	Login_ID  *string            `json:"login_id"`
	TimeStamp time.Time          `json:"timestamp"`
}
