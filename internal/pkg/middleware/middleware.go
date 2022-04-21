package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/serjyuriev/diploma-1/internal/pkg/handlers"
	"github.com/serjyuriev/diploma-1/internal/pkg/models"
)

var (
	contextKeyUID = handlers.ContextKey("user_id")

	ErrInvalidAccessToken = errors.New("invalid access token")
)

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/user/register" || r.URL.Path == "/api/user/login" {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// TODO: logging
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		headerSplitted := strings.Split(authHeader, " ")
		if headerSplitted[0] != "Bearer" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// TODO: signingKey in Config
		userID, err := parseToken(headerSplitted[1], []byte("gopherkey"))
		if err != nil {
			if errors.Is(err, ErrInvalidAccessToken) {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), contextKeyUID, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func parseToken(accessToken string, signingKey []byte) (int, error) {
	token, err := jwt.ParseWithClaims(
		accessToken,
		&models.Claims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return signingKey, nil
		},
	)

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(*models.Claims); ok && token.Valid {
		return claims.UserID, nil
	}

	return 0, ErrInvalidAccessToken
}
