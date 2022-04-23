package repository

import (
	"context"

	"github.com/serjyuriev/diploma-1/internal/pkg/models"
)

type Repository interface {
	InsertUser(ctx context.Context, user *models.User) error
	SelectUser(ctx context.Context, login string) (*models.User, error)
	InsertOrder(ctx context.Context, number string, userID int) error
	SelectOrderByNumber(ctx context.Context, number string) (*models.Order, error)
	SelectOrdersByUser(ctx context.Context, userID int) ([]*models.Order, error)
	SelectBalanceByUser(ctx context.Context, userID int) (*models.Balance, error)
	UpdateBalance(ctx context.Context, userID int, amount float64, orderID int64) error
	SelectWithdrawalsByUser(ctx context.Context, userID int) ([]*models.Order, error)
}
