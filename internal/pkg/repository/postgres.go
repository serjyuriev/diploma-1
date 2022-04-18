package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/rs/zerolog"
	"github.com/serjyuriev/diploma-1/internal/pkg/config"
	"github.com/serjyuriev/diploma-1/internal/pkg/models"

	"github.com/golang-migrate/migrate/v4"
	psql "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var (
	errNotImplemented = errors.New("method not implemented yet")
)

type postgres struct {
	cfg    config.Config
	db     *sql.DB
	logger zerolog.Logger
}

func NewPostgres(logger zerolog.Logger) (Repository, error) {
	cfg := config.GetConfig()
	logger.Debug().Msg("preparing connection to psql")
	db, err := sql.Open("pgx", cfg.DatabaseUri)
	if err != nil {
		return nil, fmt.Errorf("unable to open sql connection: %v", err)
	}

	driver, err := psql.WithInstance(db, &psql.Config{})
	if err != nil {
		return nil, fmt.Errorf("unable to create psql driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		// TODO: add this to config
		"file:///app/scripts/migrations",
		"gophermart", driver)
	if err != nil {
		return nil, fmt.Errorf("unable to create migrations client: %v", err)
	}

	m.Up()

	return &postgres{
		cfg:    cfg,
		db:     db,
		logger: logger,
	}, nil
}

func (p *postgres) InsertUser(ctx context.Context, user models.User) error {
	return errNotImplemented
}

func (p *postgres) SelectUser(ctx context.Context, user models.User) error {
	return errNotImplemented
}

func (p *postgres) InsertOrder(ctx context.Context, number, userID string) error {
	return errNotImplemented
}

func (p *postgres) SelectOrdersByUser(ctx context.Context, userID string) ([]models.Order, error) {
	return nil, errNotImplemented
}

func (p *postgres) SelectBalanceByUser(ctx context.Context, userID string) (models.Balance, error) {
	return models.Balance{}, errNotImplemented
}

func (p *postgres) UpdateBalance(ctx context.Context, userID string, amount float64) error {
	return errNotImplemented
}

func (p *postgres) SelectWithdrawalsByUser(ctx context.Context, userID string) ([]models.Order, error) {
	return nil, errNotImplemented
}
