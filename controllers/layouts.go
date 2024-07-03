// controllers/layouts.go

package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetLayouts(c *gin.Context) {
	layouts := []string{"default", "compact", "focused"}
	c.JSON(http.StatusOK, gin.H{"layouts": layouts})
}
