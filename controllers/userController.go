/*
https://www.mongodb.com/docs/drivers/go/current/fundamentals/crud/
*/

package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"go-jwt-user-auth/config"
	"go-jwt-user-auth/database"
	helper "go-jwt-user-auth/helpers"
	"go-jwt-user-auth/models"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Create variables to reference mongo collections
var usrCollection, _ = config.GetEnvVar("USER_COLLECTION")
var mntrCollection, _ = config.GetEnvVar("MONITOR_COLLECTION")
var userCollection *mongo.Collection = database.OpenCollection(database.Client, usrCollection)
var monitorCollection *mongo.Collection = database.OpenCollection(database.Client, mntrCollection)

// Used to enforce validation rules on User struct objects
var validate = validator.New()

/*
Returns hash of plaintext password string
*/
func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

/*
Compares hashed password with its possible plain text equivalent.
Used in the sign in process to compare the input password to the hash password stored in db.
Returns true if verified.
*/
func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""

	if err != nil {
		msg = fmt.Sprintf("login_id or password is incorrect")
		check = false
	}
	return check, msg
}

/*
Used in the routes/authRouter.go file to return a gin handler function to the router
Adds user to database upon signing up.
After validating that all required fields are supplied and that no such user with
the same login id already exists, the password supplied is hashed,
other User fields are generated (token, refresh token, mongo id, timestamps, etc)
and then the user is added to the db.
*/
func AddUser() gin.HandlerFunc {

	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		count, err := userCollection.CountDocuments(ctx, bson.M{"login_id": user.Login_ID})
		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the login_id"})
			return
		}

		password := HashPassword(*user.Password)
		user.Password = &password

		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this login_id already exists"})
			return
		}

		user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()
		token, refreshToken, _ := helper.GenerateAllTokens(
			*user.Login_ID,
			*user.Role, *&user.User_id)
		user.Token = &token
		user.Refresh_token = &refreshToken

		resultInsertionNumber, insertErr := userCollection.InsertOne(ctx, user)
		if insertErr != nil {
			msg := fmt.Sprintf("User item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, resultInsertionNumber)
	}

}

/*
Used in the routes/authRouter.go file to return a gin handler function to the router
Creates a user variable from the request body.
Creates a foundUser variable by looking up the stored user by user's (from request) login id.
Compares the passwords of both to authorize user.
If authorized, tokens are generated for the user and written to the user's record in the db.
*/
func SignInUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User
		var foundUser models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := userCollection.FindOne(ctx, bson.M{"login_id": user.Login_ID}).Decode(&foundUser)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "login_id or password is incorrect"})
			return
		}

		passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
		defer cancel()
		if passwordIsValid != true {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		if foundUser.Login_ID == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		}
		token, refreshToken, _ := helper.GenerateAllTokens(*foundUser.Login_ID,
			*foundUser.Role, foundUser.User_id)
		helper.UpdateAllTokens(token, refreshToken, foundUser.User_id)
		err = userCollection.FindOne(ctx, bson.M{"user_id": foundUser.User_id}).Decode(&foundUser)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, foundUser)
	}
}

/*
Used in the routes/userRouter.go file to return a gin handler function to the router
Similar to the SignIn function,
this creates a user variable from the request body and
creates a foundUser variable by looking up the login id of the user attempting to
authorize himself as a monitor.
Compares the passwords of both to authorize user.
If the attempt is from a monitor, isMontor: true is returned and a log is written
to the monitor collection in the database.const
Otherwise, isMontor: false is returned
*/
func AuthenticateMonitor() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User
		var foundUser models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := userCollection.FindOne(ctx, bson.M{"login_id": user.Login_ID}).Decode(&foundUser)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "login_id or password is incorrect"})
			return
		}

		passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
		defer cancel()
		if passwordIsValid != true {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		if foundUser.Login_ID == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		}

		isMonitor := false
		if *foundUser.Role == "monitor" {
			isMonitor = true
			now, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
			entry := models.MonitorEntry{
				ID:        primitive.NewObjectID(),
				Login_ID:  foundUser.Login_ID,
				TimeStamp: now,
			}
			_, insertErr := monitorCollection.InsertOne(ctx, entry)
			if insertErr != nil {
				println(insertErr)
				msg := "Monitor entry was not created"
				c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
				return
			}
		}
		defer cancel()
		resMap := map[string]bool{
			"isMonitor": isMonitor,
		}
		c.JSON(http.StatusOK, resMap)
	}
}

/*
Used in the routes/userRouter.go file to return a gin handler function to the router
Finds user in mongoDB by the login id, and returns the role of that user to the router
*/
func GetUserRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Param("login_id")

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var user models.User
		err := userCollection.FindOne(ctx, bson.M{"login_id": userId}).Decode(&user)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		resMap := map[string]string{
			"role": *user.Role,
		}
		c.JSON(http.StatusOK, resMap)
	}
}
