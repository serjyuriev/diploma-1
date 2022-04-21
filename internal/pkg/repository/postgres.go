package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

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
	db, err := sql.Open("pgx", cfg.DatabaseURI)
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

func (p *postgres) InsertOrder(ctx context.Context, number, userID int) error {
	return errNotImplemented
}

func (p *postgres) SelectOrdersByUser(ctx context.Context, userID int) ([]*models.Order, error) {
	p.logger.Debug().Caller().Msgf("selecting orders for user '%d'", userID)

	rows, err := p.db.Query(
		"SELECT o.number, o.status, p.amount, o.uploaded_at FROM orders AS o JOIN posting AS p ON p.order_id = o.id AND p.user_id = o.user_id WHERE o.user_id = $1 ORDER BY uploaded_at ASC;",
		userID,
	)

	if err != nil {
		p.logger.Error().Caller().Msg("unable to execute query")
		return nil, rows.Err()
	}

	orders := make([]*models.Order, 0)
	for rows.Next() {
		order := new(models.Order)
		var unixUploaded int64
		if err := rows.Scan(&order.Number, &order.Status, &order.Accrual, &unixUploaded); err != nil {
			p.logger.Error().Caller().Msg("unable to scan query result")
			return nil, err
		}
		order.UploadedAt = time.Unix(unixUploaded, 0)
		orders = append(orders, order)
	}

	if rows.Err() != nil {
		p.logger.Error().Caller().Msg("unable to execute query")
		return nil, rows.Err()
	}

	p.logger.Debug().Caller().Msgf("found %d orders for user '%d'", len(orders), userID)

	return orders, nil
}

func (p *postgres) SelectBalanceByUser(ctx context.Context, userID int) (*models.Balance, error) {
	return nil, errNotImplemented
}

func (p *postgres) UpdateBalance(ctx context.Context, userID int, amount float64) error {
	return errNotImplemented
}

func (p *postgres) SelectWithdrawalsByUser(ctx context.Context, userID int) ([]*models.Order, error) {
	return nil, errNotImplemented
}
