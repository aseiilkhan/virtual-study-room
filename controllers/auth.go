package controllers

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/aseiilkhan/virtual-study-room/config"
	"github.com/aseiilkhan/virtual-study-room/models"
)

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

var jwtSecret = []byte(os.Getenv("SECRET_KEY"))

func GenerateJWT(email string, expiration time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["email"] = email
	claims["exp"] = time.Now().Add(expiration).Unix() // Token expires in 1 hour

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func GenerateTokens(email string) (accessToken, refreshToken string, err error) {
	// Access Token (short-lived)
	accessToken, err = GenerateJWT(email, time.Hour*1) // Access token valid for 1 hour
	if err != nil {
		return "", "", err
	}

	// Refresh Token (long-lived)
	refreshToken, err = GenerateJWT(email, time.Hour*24*7) // Refresh token valid for 7 days
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func ValidateJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the token from the Authorization header
		tokenString := c.GetHeader("Authorization")

		// Check if the Authorization header is present
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			c.Abort()
			return
		}

		// Ensure that the token has the "Bearer " prefix and extract the token part
		const bearerPrefix = "Bearer "
		if len(tokenString) > len(bearerPrefix) && tokenString[:len(bearerPrefix)] == bearerPrefix {
			tokenString = tokenString[len(bearerPrefix):]
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			c.Abort()
			return
		}

		// Parse and validate the JWT token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the token's signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method")
			}
			// Return the secret key to validate the signature
			return jwtSecret, nil
		})

		// If there was an error or the token is invalid, return an unauthorized error
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
		c.Set("email", token.Claims.(jwt.MapClaims)["email"])
		fmt.Println(token.Claims.(jwt.MapClaims)["email"])
		// If the token is valid, allow the request to proceed
		c.Next()
	}
}

func Protected(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Welcome to the protected route!"})
}

func Login(c *gin.Context) {
	var creds Credentials
	var user models.User
	db := config.DB

	if err := c.BindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Check if user exists
	if err := db.Where("email = ?", creds.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Check if password is correct
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate access and refresh tokens
	accessToken, refreshToken, err := GenerateTokens(creds.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	// Return both tokens
	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func Register(c *gin.Context) {
	var creds Credentials
	var user models.User
	db := config.DB

	if err := c.BindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if err := db.Where("email = ?", creds.Email).First(&user).Error; err == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User already exists"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	db.Create(&models.User{Email: creds.Email, Password: string(hashedPassword)})
	// Store the user with the hashed password
	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

func RefreshToken(c *gin.Context) {
	var requestBody struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Parse and validate the refresh token
	token, err := jwt.Parse(requestBody.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method")
		}
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// Extract email from the refresh token claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !claims["authorized"].(bool) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		return
	}

	email := claims["email"].(string)

	// Generate a new access token
	newAccessToken, _, err := GenerateTokens(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate new access token"})
		return
	}

	// Send the new access token to the client
	c.JSON(http.StatusOK, gin.H{
		"access_token": newAccessToken,
	})
}
