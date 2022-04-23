package accrual

import (
	"context"

	"github.com/serjyuriev/diploma-1/internal/pkg/config"
	"github.com/serjyuriev/diploma-1/internal/pkg/models"
)

type Accrual interface {
	GetOrderStatus(ctx context.Context, order string) (*models.Order, error)
}

type accrual struct {
	systemURL string
}

func NewAccrualClient() Accrual {
	return &accrual{
		systemURL: config.GetConfig().AccrualSystemAddress,
	}
}

func (*accrual) GetOrderStatus(ctx context.Context, order string) (*models.Order, error) {
	return nil, nil
}
