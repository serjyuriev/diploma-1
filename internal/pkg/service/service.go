package service

import (
	"context"
	"errors"

	"github.com/rs/zerolog"
	"github.com/serjyuriev/diploma-1/internal/pkg/config"
	"github.com/serjyuriev/diploma-1/internal/pkg/models"
)

var (
	errNotImplemented = errors.New("method not implemented yet")
)

type Service interface {
	RegisterUser(ctx context.Context, user models.User) error
	CreateNewOrder(ctx context.Context, number, userID string) error
	WithdrawPoints(ctx context.Context, userID string, amount float64) error
}

type service struct {
	config config.Config
	logger zerolog.Logger
}

func NewService(logger zerolog.Logger) (Service, error) {
	return &service{
		config: config.GetConfig(),
		logger: logger,
	}, nil
}

func (svc *service) RegisterUser(ctx context.Context, user models.User) error {
	return errNotImplemented
}

func (svc *service) CreateNewOrder(ctx context.Context, number, userID string) error {
	return errNotImplemented
}

func (svc *service) WithdrawPoints(ctx context.Context, userID string, amount float64) error {
	return errNotImplemented
}
