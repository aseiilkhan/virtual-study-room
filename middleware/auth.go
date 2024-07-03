// middleware/auth.go

package middleware

import (
	"log"
	"net/http"

	"github.com/dgrijalva/jwt-go/v4"
	"github.com/gin-gonic/gin"

	"github.com/aseiilkhan/virtual-study-room/services"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			return
		}
		token, err := services.ValidateToken(tokenString)
		log.Println("Received token:", tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}
		c.Set("userId", token.Claims.(jwt.MapClaims)["userId"])
		c.Next()
	}
}
