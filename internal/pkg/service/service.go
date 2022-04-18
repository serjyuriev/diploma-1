package service

import (
	"context"
	"crypto/sha1"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/rs/zerolog"
	"github.com/serjyuriev/diploma-1/internal/pkg/config"
	"github.com/serjyuriev/diploma-1/internal/pkg/models"
	"github.com/serjyuriev/diploma-1/internal/pkg/repository"
)

var (
	errNotImplemented = errors.New("method not implemented yet")
	ErrNotRegistered  = errors.New("user with provided credentials isn't registered")
)

type Service interface {
	RegisterUser(ctx context.Context, user *models.User) (string, error)
	LoginUser(ctx context.Context, user *models.User) (string, error)
	CreateNewOrder(ctx context.Context, number, userID string) error
	WithdrawPoints(ctx context.Context, userID string, amount float64) error
}

type service struct {
	config         config.Config
	logger         zerolog.Logger
	repo           repository.Repository
	hashSalt       string
	signingKey     []byte
	expireDuration time.Duration
}

// NewService creates new instance of service structure.
func NewService(logger zerolog.Logger, repo repository.Repository) (Service, error) {
	logger.Debug().Caller().Msg("initializing service")

	return &service{
		config:         config.GetConfig(),
		logger:         logger,
		repo:           repo,
		hashSalt:       "gopher",
		signingKey:     []byte("gopherkey"),
		expireDuration: 10 * time.Minute,
	}, nil
}

// RegisterUser inserts new user credentials into database and
// tries to log user in.
func (svc *service) RegisterUser(ctx context.Context, user *models.User) (string, error) {
	dbUser := &models.User{Login: user.Login}

	pwd := sha1.New()
	pwd.Write([]byte(user.Password))
	pwd.Write([]byte(svc.hashSalt))
	dbUser.Password = fmt.Sprintf("%x", pwd.Sum(nil))

	if err := svc.repo.InsertUser(ctx, dbUser); err != nil {
		svc.logger.Error().Caller().Msg("unable to insert user into database")
		return "", err
	}

	return svc.LoginUser(ctx, user)
}

// LoginUser gathers user credentials from database
// and compares them with provided by caller,
// returning error in case they don't match.
func (svc *service) LoginUser(ctx context.Context, user *models.User) (string, error) {
	svc.logger.Debug().Caller().Msgf("logging in user '%s'", user.Login)

	pwd := sha1.New()
	pwd.Write([]byte(user.Password))
	pwd.Write([]byte(svc.hashSalt))
	user.Password = fmt.Sprintf("%x", pwd.Sum(nil))

	dbUser, err := svc.repo.SelectUser(ctx, user.Login)
	if err != nil {
		svc.logger.Error().Caller().Msg("unable to select user from database")
		return "", err
	}

	if dbUser.Login == user.Login && dbUser.Password == user.Password {
		token := jwt.NewWithClaims(
			jwt.SigningMethodHS256,
			models.Claims{
				StandardClaims: jwt.StandardClaims{
					ExpiresAt: time.Now().Add(svc.expireDuration).Unix(),
					IssuedAt:  time.Now().Unix(),
				},
				UserID: user.ID,
			},
		)

		svc.logger.Debug().Caller().Msgf("user '%s' logged in", user.Login)
		return token.SignedString(svc.signingKey)
	}

	svc.logger.Info().Caller().Msgf("unsuccessful attempt to login for '%s'", user.Login)
	return "", ErrNotRegistered
}

func (svc *service) CreateNewOrder(ctx context.Context, number, userID string) error {
	return errNotImplemented
}

func (svc *service) WithdrawPoints(ctx context.Context, userID string, amount float64) error {
	return errNotImplemented
}
