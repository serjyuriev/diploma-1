package repository

import (
	"context"

	"github.com/serjyuriev/diploma-1/internal/pkg/models"
)

type Repository interface {
	InsertUser(ctx context.Context, user models.User) error
	SelectUser(ctx context.Context, user models.User) error
	InsertOrder(ctx context.Context, number, userID string) error
	SelectOrdersByUser(ctx context.Context, userID string) ([]models.Order, error)
	SelectBalanceByUser(ctx context.Context, userID string) (models.Balance, error)
	UpdateBalance(ctx context.Context, userID string, amount float64) error
	SelectWithdrawalsByUser(ctx context.Context, userID string) ([]models.Order, error)
}
