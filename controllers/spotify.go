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
	"time"

	"github.com/aseiilkhan/virtual-study-room/config"
	"github.com/aseiilkhan/virtual-study-room/models"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)

func GetSpotifyAuthLogin(c *gin.Context) {
	scope := "streaming user-read-email user-read-private user-read-playback-state user-modify-playback-state user-library-read user-library-modify user-read-currently-playing"
	state := generateRandomString(16)

	// Get the user email from the context, ensure it's a string
	userEmail, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email not found in context"})
		return
	}

	emailStr, ok := userEmail.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid email type"})
		return
	}

	// Create a new instance of the state model and save it to the DB
	stateCache := models.State{
		State: state,
		Email: emailStr,
	}

	// Insert the state and email into the database
	db := config.DB
	if err := db.Create(&stateCache).Error; err != nil {
		fmt.Println("Error saving state:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save state"})
		return
	}

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
		"redirect_uri":  {"https://virtual-study-room-c7879980fd07.herokuapp.com/api/spotify/auth/callback"}, // Adjust if your frontend is on a different port
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

	db := config.DB
	var stateCache models.State
	db_search_result := db.Model(&stateCache).Where("state = ?", c.Query("state")).First(&stateCache)
	if db_search_result.Error != nil {
		fmt.Println(db_search_result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch state"})
		return
	}

	userEmail := stateCache.Email
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
		"redirect_uri": "https://virtual-study-room-c7879980fd07.herokuapp.com/api/spotify/auth/callback", // Adjust if your frontend is on a different port
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

	fmt.Println(resp.String())
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

	expires_in, ok := result["expires_in"].(float64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid token response format"})
		return
	}

	expiry_time := int(expires_in) + int(time.Now().Unix())

	refreshToken, ok := result["refresh_token"].(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid token response format"})
		return
	}
	// Update user record with access token

	var user models.User
	db_search_resultt := db.Model(&user).Where("email = ?", userEmail).Update("spotify_token", accessToken).Update("spotify_refresh_token", refreshToken).Update("spotify_token_expires_at", expiry_time) // Update this line
	if db_search_resultt.Error != nil {
		fmt.Println(db_search_result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save access token"})
		return
	}
	// c.SetCookie("spotify_token", accessToken, 3600, "/", "localhost", false, true)
	// c.SetCookie("refresh_token", result["refresh_token"].(string), 3600, "/", "localhost", false, true)
	c.Redirect(http.StatusFound, "http://localhost:3000")
}

func GetSpotifyAuthToken(c *gin.Context) {
	db := config.DB

	var user models.User
	userEmail, exists := c.Get("email")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email not found in context"})
		return
	}

	if err := db.Where("email = ?", userEmail).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}

	if user.SpotifyToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Access token not found for user"})
		return
	}

	if user.SpotifyTokenExpiresAt < (time.Now().Unix()) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Access token expired"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": user.SpotifyToken, "sptotify_token_expires_at": user.SpotifyTokenExpiresAt, "refresh_token": user.SpotifyRefreshToken})
}
func GetSpotifyRefreshToken(c *gin.Context) {
	db := config.DB
	var user models.User

	userEmail, exists := c.Get("email")
	fmt.Println(userEmail)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email not found in context"})
		return
	}

	fmt.Println(userEmail)
	if err := db.Where("email = ?", userEmail).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}

	if user.SpotifyRefreshToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token not found for user"})
		return
	}

	spotifyClientID := os.Getenv("SPOTIFY_CLIENT_ID")
	spotifyClientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")

	if spotifyClientID == "" || spotifyClientSecret == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Spotify credentials not set"})
		return
	}

	authHeader := base64.StdEncoding.EncodeToString([]byte(spotifyClientID + ":" + spotifyClientSecret))

	// Prepare the request body
	requestBody := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": user.SpotifyRefreshToken,
	}

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

	expires_in, ok := result["expires_in"].(float64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid token response format"})
		return
	}
	expiry_time := int(expires_in) + int(time.Now().Unix())

	refreshToken, ok := result["refresh_token"].(string)
	if !ok {
		refreshToken = user.SpotifyRefreshToken
	}

	db_search_result := db.Model(&user).Where("email = ?", userEmail).Update("spotify_token", accessToken).Update("spotify_refresh_token", refreshToken).Update("spotify_token_expires_at", expiry_time)
	if db_search_result.Error != nil {
		fmt.Println(db_search_result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save access token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": accessToken, "sptotify_token_expires_at": expiry_time, "refresh_token": refreshToken})
}
func generateRandomString(length int) string {
	text := ""
	possible := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	for i := 0; i < length; i++ {
		text += string(possible[rand.Intn(len(possible))])
	}
	return text
}
