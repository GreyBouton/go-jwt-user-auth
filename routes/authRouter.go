/*
Houses the routes where username and password are required (authorization)
*/
package routes

import (
	controller "go-jwt-user-auth/controllers"

	"github.com/gin-gonic/gin"
)

func AuthRoutes(router *gin.Engine) {
	router.POST("users/adduser", controller.AddUser())
	router.POST("users/signin", controller.SignInUser())
}
