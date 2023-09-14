/*
Houses (protected) routes that are accessible to authorized users and thus
enforces authentication (JWT)
*/

package routes

import (
	controller "go-jwt-user-auth/controllers"
	"go-jwt-user-auth/middleware"

	"github.com/gin-gonic/gin"
)

func UserRoutes(router *gin.Engine) {

	router.Use(middleware.Authenticate()) // Enforces JWT authenticaton for the following routes
	router.GET("/users/role/:login_id", controller.GetUserRole())
	router.POST("/users/authenticatemonitor", controller.AuthenticateMonitor())
}
