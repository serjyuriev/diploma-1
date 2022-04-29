package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/rs/zerolog"
	"github.com/serjyuriev/diploma-1/internal/app/handlers"
	"github.com/serjyuriev/diploma-1/internal/pkg/config"
	"github.com/serjyuriev/diploma-1/internal/pkg/models"
)

var (
	contextKeyUID = handlers.ContextKey("user_id")

	ErrInvalidAccessToken = errors.New("invalid access token")
)

type Middleware interface {
	Auth(next http.Handler) http.Handler
}

type middleware struct {
	cfg    config.Config
	logger zerolog.Logger
}

func NewMiddleware(logger zerolog.Logger) Middleware {
	return &middleware{
		cfg:    config.GetConfig(),
		logger: logger,
	}
}

func (m *middleware) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/user/register" || r.URL.Path == "/api/user/login" {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.logger.
				Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("code", http.StatusUnauthorized).
				Msg("request has no authorization headers")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		headerSplitted := strings.Split(authHeader, " ")
		if headerSplitted[0] != "Bearer" {
			m.logger.
				Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("code", http.StatusUnauthorized).
				Msg("authorization method is not valid")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userID, err := parseToken(headerSplitted[1], []byte(m.cfg.SigningKey))
		if err != nil {
			if errors.Is(err, ErrInvalidAccessToken) {
				m.logger.
					Info().
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Int("code", http.StatusUnauthorized).
					Msg("provided token is not valid")
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			m.logger.
				Err(err).
				Caller().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("code", http.StatusInternalServerError).
				Msg("unexpected error occured while processing JWT token")
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		m.logger.
			Debug().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("user_id", userID).
			Msg("user was authorized")
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
