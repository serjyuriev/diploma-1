package repository

import (
	"context"
	"database/sql"
	"errors"

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

// NewPostgres creates new instance of PostgreSQL implementation
// of Repository interface.
func NewPostgres(logger zerolog.Logger) (Repository, error) {
	cfg := config.GetConfig()
	logger.Debug().Caller().Msg("preparing connection to psql")
	db, err := sql.Open("pgx", cfg.DatabaseUri)
	if err != nil {
		logger.Error().Caller().Msg("unable to open sql connection")
		return nil, err
	}

	driver, err := psql.WithInstance(db, &psql.Config{})
	if err != nil {
		logger.Error().Caller().Msg("unable to create psql driver")
		return nil, err
	}

	m, err := migrate.NewWithDatabaseInstance(
		// TODO: add this to config
		"file:///app/scripts/migrations",
		"gophermart", driver)
	if err != nil {
		logger.Error().Caller().Msg("unable to create migrations client")
		return nil, err
	}

	m.Up()

	return &postgres{
		cfg:    cfg,
		db:     db,
		logger: logger,
	}, nil
}

// InsertUser inserts provided user information into users table.
func (p *postgres) InsertUser(ctx context.Context, user *models.User) error {
	p.logger.Debug().Caller().Msgf("inserting user '%s' in db", user.Login)

	if _, err := p.db.Exec(
		"INSERT INTO users (login, password) VALUES ($1, $2)",
		user.Login,
		user.Password,
	); err != nil {
		p.logger.Error().Caller().Msg("unable to execute query")
		return err
	}

	p.logger.Debug().Caller().Msgf("user '%s' was inserted", user.Login)

	return nil
}

// SelectUser gathers user information from users table based on provided login.
func (p *postgres) SelectUser(ctx context.Context, login string) (*models.User, error) {
	p.logger.Debug().Caller().Msgf("selecting user with login '%s'", login)

	rows := p.db.QueryRow(
		"SELECT id, login, password FROM users WHERE login = $1",
		login,
	)
	if rows.Err() != nil {
		p.logger.Error().Caller().Msg("unable to execute query")
		return nil, rows.Err()
	}

	user := new(models.User)
	if err := rows.Scan(&user.ID, &user.Login, &user.Password); err != nil {
		p.logger.Error().Caller().Msg("unable to scan query result")
		return nil, err
	}

	p.logger.Debug().Caller().Msgf("found user with login '%s'", user.Login)

	return user, nil
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
