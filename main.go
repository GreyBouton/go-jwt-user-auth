package main

import (
	config "go-jwt-user-auth/config"

	routes "go-jwt-user-auth/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	/*
		Any time the _ is used for assignment, it is because the function on the
		right side returns multiple items and we store the uneeded items in the _
		to throw it away.
	*/
	port, _ := config.GetEnvVar("PORT")

	// If env still empty due to no value in .env file, default to 8000
	if port == "" {
		port = "8000"
	}

	// Instantiate a router to be used to direct HTTP requests to the code that handles them.
	router := gin.New()
	// Attach middleware logger to be used for every request
	router.Use(gin.Logger())

	/*
		The following two lines are essentially doing this:
		router.<GET/POST/PATCH/ETC>("path/to/endpoint",functionThatPerformsBusinessLogic())
		If you investicate the below functions, you will se that they just wrap what is commented above,
		splitting functions between ones that reqyure authorization vs ones that require authentication (JWT)
	*/
	routes.AuthRoutes(router)
	routes.UserRoutes(router)

	/*
		Atttaches router to HTTP server and starts listening to and serving
		requests on the specified port
	*/
	router.Run(":" + port)
}
