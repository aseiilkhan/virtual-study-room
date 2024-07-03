// services/auth.go

package services

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aseiilkhan/virtual-study-room/models"

	"github.com/dgrijalva/jwt-go/v4"
	"golang.org/x/crypto/bcrypt"
)

var SecretKey = os.Getenv("SECRET")

func GenerateToken(user models.User) (string, error) {
	claims := jwt.MapClaims{
		"userId": user.ID,
		"exp":    time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	fmt.Println(token)
	return token.SignedString([]byte(SecretKey))
}

func ValidateToken(tokenString string) (*jwt.Token, error) {
	parts := strings.Split(tokenString, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, fmt.Errorf("invalid token format")
	}
	tokenString = parts[1]
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})
	if err != nil || !token.Valid {
		log.Println("Error validating token in serivces/auth.go :", err)
		return nil, err
	}

	fmt.Println(token)
	return token, nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
