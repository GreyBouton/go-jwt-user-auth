package middleware

import (
	"net/http"

	helper "go-jwt-user-auth/helpers"

	"github.com/gin-gonic/gin"
)

/*
Used in the routes/userRouter.go file to return a gin handler function to the router
This function enforces JWT authentication in order for a user to access protected routes.
If the token is null or wrong, the function ensures that authenitcated
handlers are prevented from being called.
gin.context.set:
https://golang.hotexamples.com/examples/github.com.gin-gonic.gin/Context/Set/golang-context-set-method-examples.html
*/
func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientToken := c.Request.Header.Get("token")
		if clientToken == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "No Authorization header provided"})
			c.Abort()
			return
		}

		claims, err := helper.ValidateToken(clientToken)
		if err != "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			c.Abort()
			return
		}
		// Adding login_id, uid, and role to context
		c.Set("login_id", claims.Login_ID)
		c.Set("uid", claims.Uid)
		c.Set("role", claims.Role)
		c.Next()
	}
}
