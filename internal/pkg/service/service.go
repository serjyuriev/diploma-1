package service

import (
	"context"
	"errors"

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
	RegisterUser(ctx context.Context, user models.User) error
	LoginUser(ctx context.Context, user models.User) error
	CreateNewOrder(ctx context.Context, number, userID string) error
	WithdrawPoints(ctx context.Context, userID string, amount float64) error
}

type service struct {
	config config.Config
	logger zerolog.Logger
	repo   repository.Repository
}

// NewService creates new instance of service structure.
func NewService(logger zerolog.Logger, repo repository.Repository) (Service, error) {
	logger.Debug().Caller().Msg("initializing service")

	return &service{
		config: config.GetConfig(),
		logger: logger,
		repo:   repo,
	}, nil
}

// RegisterUser inserts new user credentials into database and
// tries to log user in.
func (svc *service) RegisterUser(ctx context.Context, user models.User) error {
	if err := svc.repo.InsertUser(ctx, user); err != nil {
		svc.logger.Error().Caller().Msg("unable to insert user into database")
		return err
	}

	if err := svc.LoginUser(ctx, user); err != nil {
		if errors.Is(err, ErrNotRegistered) {
			return err
		} else {
			svc.logger.Error().Caller().Msg("unable to login user")
			return err
		}
	}

	return nil
}

// LoginUser gathers user credentials from database
// and compares them with provided by caller,
// returning error in case they don't match.
func (svc *service) LoginUser(ctx context.Context, user models.User) error {
	svc.logger.Debug().Caller().Msgf("logging in user '%s'", user.Login)

	dbUser, err := svc.repo.SelectUser(ctx, user.Login)
	if err != nil {
		svc.logger.Error().Caller().Msg("unable to select user from database")
		return err
	}

	if dbUser.Login == user.Login && dbUser.Password == user.Password {
		svc.logger.Debug().Caller().Msgf("user '%s' logged in", user.Login)
		return nil
	}

	svc.logger.Info().Caller().Msgf("unsuccessful attempt to login for '%s'", user.Login)
	return ErrNotRegistered
}

func (svc *service) CreateNewOrder(ctx context.Context, number, userID string) error {
	return errNotImplemented
}

func (svc *service) WithdrawPoints(ctx context.Context, userID string, amount float64) error {
	return errNotImplemented
}
