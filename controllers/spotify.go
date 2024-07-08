package controllers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"

	"github.com/aseiilkhan/virtual-study-room/config"
	"github.com/aseiilkhan/virtual-study-room/models"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)

var access_tokens []string

func GetSpotifyAuthLogin(c *gin.Context) {
	scope := "streaming user-read-email user-read-private user-read-playback-state user-modify-playback-state user-library-read user-library-modify"
	state := generateRandomString(16)

	// Get your client ID from the environment
	spotifyClientID := os.Getenv("SPOTIFY_CLIENT_ID")
	if spotifyClientID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "SPOTIFY_CLIENT_ID not set"})
		return
	}

	authQueryParams := url.Values{
		"response_type": {"code"},
		"client_id":     {spotifyClientID},
		"scope":         {scope},
		"redirect_uri":  {"http://localhost:8080/api/spotify/auth/callback"}, // Adjust if your frontend is on a different port
		"state":         {state},
	}

	authURL := "https://accounts.spotify.com/authorize/?" + authQueryParams.Encode()
	log.Println(authURL)
	c.JSON(http.StatusOK, gin.H{"auth_url": authURL})
}

// Callback handles the Spotify authorization callback
func GetSpotifyAuthCallback(c *gin.Context) {
	code := c.Query("code")
	log.Println(code)

	// Get your client ID and secret from the environment
	spotifyClientID := os.Getenv("SPOTIFY_CLIENT_ID")
	spotifyClientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")

	if spotifyClientID == "" || spotifyClientSecret == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Spotify credentials not set"})
		return
	}

	// Construct the basic authorization header
	authHeader := base64.StdEncoding.EncodeToString([]byte(spotifyClientID + ":" + spotifyClientSecret))

	// Prepare the request body
	requestBody := map[string]string{
		"code":         code,
		"redirect_uri": "http://localhost:8080/api/spotify/auth/callback", // Adjust if your frontend is on a different port
		"grant_type":   "authorization_code",
	}

	// Create a Resty client
	client := resty.New()
	// Make the POST request to the Spotify token endpoint
	resp, err := client.R().
		SetFormData(requestBody).
		SetHeader("Authorization", "Basic "+authHeader).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		Post("https://accounts.spotify.com/api/token")

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error getting token: %v", err)})
		return
	}

	log.Println("RESP IS " + resp.String())
	if resp.StatusCode() != http.StatusOK {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get token"})
		return
	}

	var result map[string]interface{} // Assuming the Spotify response is JSON
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error parsing token response: %v", err)})
		return
	}

	accessToken, ok := result["access_token"].(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid token response format"})
		return
	}
	// userEmail, exists := c.Get("email")
	// if !exists {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "User email not found in context"})
	// 	return
	// }
	userEmail := "testuser@example.com"
	// Update user record with access token
	db, err := config.ConnectDatabase()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to database"})
		return
	}
	var user models.User
	db_search_result := db.Model(&user).Where("email = ?", userEmail).Update("spotify_token", accessToken) // Update this line
	if db_search_result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save access token"})
		return
	}
	c.Redirect(http.StatusFound, "http://localhost:3000?access_token="+accessToken) // Redirect to frontend
}

func GetSpotifyAuthToken(c *gin.Context) {
	if len(access_tokens) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Access token not available"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"access_token": "BQAa1Hq-cdTqdVwrHNvKMpzoB-XkeXViPU5Gl8pD5w--7a1jMKcw4EYQLhf_6fQpQk2b713vH89ojCZUQhYeiWSpv11xorTykDcqzkL3bo-2PC0QZSsO0tSPOVuYmuxMJrJHT8kpWoCfRWf-rm6ZJAioOKx8CgClF1WlqEVDO7uDx6O6NuT6DU60hHw_lij-hE1N_3oxAARGEuRHRjaSsQqNnZkwo7Qr",
	})
}

func generateRandomString(length int) string {
	text := ""
	possible := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	for i := 0; i < length; i++ {
		text += string(possible[rand.Intn(len(possible))])
	}
	return text
}
