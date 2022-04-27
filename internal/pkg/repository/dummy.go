package repository

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/serjyuriev/diploma-1/internal/pkg/models"
)

type dummy struct {
}

// NewPostgres creates new instance of PostgreSQL implementation
// of Repository interface.
func NewDummy(logger zerolog.Logger) (Repository, error) {
	return &dummy{}, nil
}

// InsertUser inserts provided user information into users table.
func (d *dummy) InsertUser(ctx context.Context, user *models.User) error {
	return errNotImplemented
}

// SelectUser gathers user information from users table based on provided login.
func (d *dummy) SelectUser(ctx context.Context, login string) (*models.User, error) {
	return nil, errNotImplemented
}

// InsertOrder inserts new order info into orders table.
func (d *dummy) InsertOrder(ctx context.Context, number string, userID int) (int64, error) {
	return 0, errNotImplemented
}

// SelectOrderByNumber selects id of order with provided number
// from orders table.
func (d *dummy) SelectOrderByNumber(ctx context.Context, number string) (*models.Order, error) {
	return nil, errNotImplemented
}

// SelectOrdersByUser gathers number, status, accrual
// and time of uploaded of user with provided ID.
func (d *dummy) SelectOrdersByUser(ctx context.Context, userID int) ([]*models.Order, error) {
	return nil, errNotImplemented
}

func (d *dummy) UpdateOrderStatus(ctx context.Context, number string, order *models.Order) error {
	return errNotImplemented
}

// SelectBalanceByUser calculates amount of points currently
// awailable to user and amount of already withdrawn points.
func (d *dummy) SelectBalanceByUser(ctx context.Context, userID int) (*models.Balance, error) {
	return nil, errNotImplemented
}

// InsertWithdrawal insert amount of withdrawn points into
// posting table.
func (d *dummy) InsertWithdrawal(ctx context.Context, userID int, amount float64, orderID int64) error {
	return errNotImplemented
}

// InsertAccrual insert amount of added points into
// posting table.
func (d *dummy) InsertAccrual(ctx context.Context, userID int, amount float64, orderID int64) error {
	return errNotImplemented
}

// SelectWithdrawalsByUser gather order's number, sum and time of processing
// for provided user ID.
func (d *dummy) SelectWithdrawalsByUser(ctx context.Context, userID int) ([]*models.Order, error) {
	return nil, errNotImplemented
}
