package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/wbrijesh/identity/internal/models"
)

var jwtSecret = []byte("your_secret_key_here") // Replace with a secure secret key

func GenerateAdminJWT(admin *models.ResponseAdmin) (string, error) {
	claims := jwt.MapClaims{
		"id":    admin.ID,
		"email": admin.Email,
		"role":  "admin",
		"exp":   time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func GenerateUserJWT(user *models.ResponseUser) (string, error) {
	claims := jwt.MapClaims{
		"id":             user.ID,
		"email":          user.Email,
		"application_id": user.ApplicationID,
		"role":           "user",
		"exp":            time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
