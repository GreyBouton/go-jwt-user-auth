package helper

import (
	"context"
	"fmt"
	"log"
	"time"

	"go-jwt-user-auth/config"
	"go-jwt-user-auth/database"

	jwt "github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SignedDetails struct {
	Login_ID string
	Uid      string
	Role     string
	jwt.StandardClaims
}

// Create collection variable to interact with users collection in db
var collection, _ = config.GetEnvVar("USER_COLLECTION")
var userCollection *mongo.Collection = database.OpenCollection(database.Client, collection)

var SECRET_KEY, _ = config.GetEnvVar("SECRET_KEY")

/*
Generates JWT tokens (token and refresh token) from login id, role, and uid
https://pkg.go.dev/github.com/golang-jwt/jwt#section-readme
*/
func GenerateAllTokens(loginID string, role string, uid string) (signedToken string, signedRefreshToken string, err error) {
	/*
		Create a signedDetails object.
		'claims' is used because in JWT terminology,
		claims are just pieces of information about the auhtenticated subject
	*/
	claims := &SignedDetails{
		Login_ID: loginID,
		Uid:      uid,
		Role:     role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(), //24 hour expiration
		},
	}

	refreshClaims := &SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(168)).Unix(), //168 hour expiration
		},
	}

	// Creating the two tokens
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))

	if err != nil {
		log.Panic(err)
		return
	}

	return token, refreshToken, err
}

/*
Validates that the token from the request header matches the token generated.
Furthermore, also checks that the token is not expired.
If any checks fail, error message is returned
*/
func ValidateToken(signedToken string) (claims *SignedDetails, msg string) {
	//https://golang.hotexamples.com/examples/github.com.dgrijalva.jwt-go/-/ParseWithClaims/golang-parsewithclaims-function-examples.html#:~:text=token%2C%20err%20%3A%3D%20jwt.ParseWithClaims(cookie.Value%2C%20%26u.TokenClaims%7B%7D%2C%20func(token%20*jwt.Token)%20(interface%7B%7D%2C%20error)%20%7B
	token, err := jwt.ParseWithClaims(
		signedToken,
		&SignedDetails{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		},
	)

	if err != nil {
		msg = err.Error()
		return
	}

	claims, ok := token.Claims.(*SignedDetails)
	if !ok {
		msg = fmt.Sprintf("the token is invalid")
		msg = err.Error()
		return
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		msg = fmt.Sprintf("token is expired")
		msg = err.Error()
		return
	}
	return claims, msg
}

/*
Update record with token and refresh toekn in db corresonding to user id
*/
func UpdateAllTokens(signedToken string, signedRefreshToken string, userId string) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	var updateObj primitive.D

	updateObj = append(updateObj, bson.E{"token", signedToken})
	updateObj = append(updateObj, bson.E{"refresh_token", signedRefreshToken})

	Updated_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updateObj = append(updateObj, bson.E{"updated_at", Updated_at})

	upsert := true
	filter := bson.M{"user_id": userId}
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}

	_, err := userCollection.UpdateOne(
		ctx,
		filter,
		bson.D{
			{"$set", updateObj},
		},
		&opt,
	)

	defer cancel()

	if err != nil {
		log.Panic(err)
		return
	}
	return
}
