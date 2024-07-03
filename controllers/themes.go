// controllers/themes.go

package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetThemes(c *gin.Context) {
	themes := []string{"light", "dark", "nature", "cyberpunk"}
	c.JSON(http.StatusOK, gin.H{"themes": themes})
}
