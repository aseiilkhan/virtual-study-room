package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/aseiilkhan/virtual-study-room/config"
	"github.com/aseiilkhan/virtual-study-room/models"
)

// Get /api/preferences
func GetPreferences(c *gin.Context) {
	// Connect to the database, retry up to 3 times
	db, err := config.ConnectDatabase()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to database"})
		return
	}

	var preferences models.Preferences
	userID := c.Query("userId")

	// Find preferences of userID
	if result := db.Where("user_id = ?", userID).First(&preferences); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, preferences)
}

// PUT /api/preferences
func UpdatePreferences(c *gin.Context) {
	// Connect to the database, retry up to 3 times
	db, err := config.ConnectDatabase()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to database"})
		return
	}

	userID := c.Query("userId")
	var existingPrefs models.Preferences
	// Find existing preferences of userID
	if err := db.Where("user_id = ?", userID).First(&existingPrefs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch preferences"})
		return
	}

	// Get new preferences from request body
	var newPreferences models.Preferences
	if err := c.BindJSON(&newPreferences); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "error binding json? " + err.Error()})
		return
	}

	// Update only new preferences
	existingPrefs.Theme = updatePreferenceIfChanged(existingPrefs.Theme, newPreferences.Theme)
	existingPrefs.Layout = updatePreferenceIfChanged(existingPrefs.Layout, newPreferences.Layout)

	// Save updated preferences to the database
	if err := db.Save(&existingPrefs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update preferences"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Preferences updated successfully"})
}

func updatePreferenceIfChanged(originalValue, newValue string) string {
	if newValue != "" {
		return newValue
	}
	return originalValue
}
