package auth

import (
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var refreshSecretKey = []byte("my_refresh_secret_key")

type RefreshTokenClaims struct {
	ApplicationID string `json:"app_id"`
	jwt.RegisteredClaims
}

var accessSecretKey = []byte("my_access_secret_key")

type AccessTokenClaims struct {
	ApplicationID string `json:"app_id"`
	jwt.RegisteredClaims
}

func GenerateRefreshToken(applicationID string) (string, error) {
	claims := RefreshTokenClaims{
		ApplicationID: applicationID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour * 7)), // Valid for 7 days
		},
	}

	// Create a new token object
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the refresh secret key
	refreshToken, err := token.SignedString(refreshSecretKey)
	if err != nil {
		return "", err
	}

	return refreshToken, nil
}

func GenerateAccessTokenFromRefreshToken(refreshToken string) (string, error) {
	// Parse and validate the refresh token
	token, err := jwt.ParseWithClaims(refreshToken, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return refreshSecretKey, nil
	})
	if err != nil || !token.Valid {
		return "", fmt.Errorf("invalid refresh token")
	}

	// Extract the application ID from the refresh token
	claims, ok := token.Claims.(*RefreshTokenClaims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	// Create a new access token with a shorter expiry time (e.g., 15 minutes)
	accessTokenClaims := AccessTokenClaims{
		ApplicationID: claims.ApplicationID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		},
	}

	// Create a new token object
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)

	// Sign the token with the access secret key
	signedAccessToken, err := accessToken.SignedString(accessSecretKey)
	if err != nil {
		return "", err
	}

	return signedAccessToken, nil
}

func ValidateAccessToken(accessToken string) (string, error) {
	// Parse and validate the access token
	token, err := jwt.ParseWithClaims(accessToken, &AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return accessSecretKey, nil
	})
	if err != nil || !token.Valid {
		return "", fmt.Errorf("invalid access token")
	}

	// Extract the application ID from the access token
	claims, ok := token.Claims.(*AccessTokenClaims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	// Return the application ID
	return claims.ApplicationID, nil
}

func TestRefreshTokens() {
	// Generate a refresh token
	refreshToken, err := GenerateRefreshToken("app_12345")
	if err != nil {
		log.Fatal("Error generating refresh token:", err)
	}
	fmt.Println("Refresh Token:", refreshToken)

	// Generate an access token from the refresh token
	accessToken, err := GenerateAccessTokenFromRefreshToken(refreshToken)
	if err != nil {
		log.Fatal("Error generating access token:", err)
	}
	fmt.Println("Access Token:", accessToken)

	// Validate the access token and get the application ID
	appID, err := ValidateAccessToken(accessToken)
	if err != nil {
		log.Fatal("Error validating access token:", err)

	}
	fmt.Println("Valid Access Token belongs to Application ID:", appID)
}
