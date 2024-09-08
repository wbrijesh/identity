package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/wbrijesh/identity/internal/auth"
)

func AcessTokenAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || strings.ToLower(bearerToken[0]) != "bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := bearerToken[1]
		adminID, err := auth.ValidateAccessToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Add the admin to the request context
		ctx := context.WithValue(r.Context(), "adminID", adminID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
