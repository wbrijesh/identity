package auth

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/wbrijesh/identity/internal/models"
)

func ValidateAdminJWT(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if claims["role"] != "admin" {
			return "", errors.New("token is not for an admin")
		}

		admin := &models.Admin{
			ID:    claims["id"].(string),
			Email: claims["email"].(string),
		}
		return admin.ID, nil
	}

	return "", errors.New("invalid token")
}

func ValidateUserJWT(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if claims["role"] != "user" {
			return "", errors.New("token is not for a user")
		}

		user := &models.User{
			ID:            claims["id"].(string),
			Email:         claims["email"].(string),
			ApplicationID: claims["application_id"].(string),
		}
		return user.ID, nil
	}

	return "", errors.New("invalid token")
}
