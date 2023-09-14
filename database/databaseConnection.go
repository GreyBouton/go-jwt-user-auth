package database

import (
	"context"
	"fmt"
	config "go-jwt-user-auth/config"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Returns a MongoDB client
func DBinstance() *mongo.Client {
	MongoDb, _ := config.GetEnvVar("MONGODB_URL")

	// Create Mongo client
	client, err := mongo.NewClient(options.Client().ApplyURI(MongoDb))
	if err != nil {
		log.Fatal(err)
	}

	// ctx is used to manage the Connection such that it purposely errors if it takes too long
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	/*
	 Using defer delays the execution of cancel() to the end of the function
	 that envelops it, in this case the DBinstance() func
	*/
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB!")

	return client
}

/*
Capitalized first letter in Client denotes that this object is meant to be
imported by other packages
*/
var Client *mongo.Client = DBinstance()

// returns a collection object (MongoDB equivalent of a table in RDS)
func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	DatabaseName, _ := config.GetEnvVar("DATABASE_NAME")

	var collection *mongo.Collection = client.Database(DatabaseName).Collection(collectionName)
	return collection
}
