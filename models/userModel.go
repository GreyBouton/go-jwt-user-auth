package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

/*
Model for user object. The json tags help convert between json and Go as
Go does not understand json and Mongo does not understand Go.
Data in Mongo will look like the json attributes.
Validate tags are used to enforce requirements on applied fields
*/
type User struct {
	ID            primitive.ObjectID `bson:"_id"`
	Login_ID      *string            `json:"login_id" validate:"required,min=7"`
	Password      *string            `json:"Password" validate:"required,min=8"`
	Role          *string            `json:"role" validate:"required,eq=admin|eq=recorder|eq=monitor|eq=reviewer"`
	Token         *string            `json:"token"`
	Refresh_token *string            `json:"refresh_token"`
	Created_at    time.Time          `json:"created_at"`
	Updated_at    time.Time          `json:"updated_at"`
	User_id       string             `json:"user_id"`
}
