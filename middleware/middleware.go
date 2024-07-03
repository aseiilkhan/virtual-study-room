package middleware

import (
	"log"

	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 {
			errorToPrint := c.Errors.ByType(gin.ErrorTypePrivate).Last()
			log.Println(errorToPrint.Error())
		}
	}
}
