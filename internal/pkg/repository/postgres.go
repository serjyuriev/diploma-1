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

	// driver, err := psql.WithInstance(db, &psql.Config{})
	// if err != nil {
	// 	logger.Error().Caller().Msg("unable to create psql driver")
	// 	return nil, err
	// }

	// m, err := migrate.NewWithDatabaseInstance(
	// 	// TODO: add this to config
	// 	"file:///app/scripts/migrations",
	// 	strings.Split(cfg.DatabaseURI, "/")[3],
	// 	driver,
	// )
	// if err != nil {
	// 	logger.Error().Caller().Msg("unable to create migrations client")
	// 	return nil, err
	// }

	// m.Up()

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

// InsertOrder inserts new order info into orders table.
func (p *postgres) InsertOrder(ctx context.Context, number string, userID int) (int64, error) {
	p.logger.Debug().Caller().Msgf("inserting order '%s' in database", number)

	row := p.db.QueryRowContext(
		ctx,
		"INSERT INTO orders (number, user_id, status, uploaded_at) VALUES ($1, $2, $3, $4) RETURNING id",
		number,
		userID,
		"NEW",
		time.Now().Unix(),
	)
	if row.Err() != nil {
		p.logger.Error().Caller().Msg("unable to execute query")
		return 0, row.Err()
	}

	var id int64
	if err := row.Scan(&id); err != nil {
		p.logger.Error().Caller().Msg("unable to scan query result")
		return 0, row.Err()
	}

	p.logger.Debug().Caller().Msgf("order '%s' was inserted", number)

	return id, nil
}

// SelectOrderByNumber selects id of order with provided number
// from orders table.
func (p *postgres) SelectOrderByNumber(ctx context.Context, number string) (*models.Order, error) {
	p.logger.Debug().Caller().Msgf("selecting order with number '%s'", number)

	row := p.db.QueryRowContext(
		ctx,
		"SELECT id, user_id FROM orders WHERE number = $1;",
		number,
	)

	order := new(models.Order)
	if err := row.Scan(&order.ID, &order.UserID); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			p.logger.Error().Caller().Msg("unable to scan query result")
		}
		return nil, err
	}
	if row.Err() != nil {
		p.logger.Error().Caller().Msg("unable to execute query")
		return nil, row.Err()
	}

	return order, nil
}

// SelectOrdersByUser gathers number, status, accrual
// and time of uploaded of user with provided ID.
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

func (p *postgres) UpdateOrderStatus(ctx context.Context, number string, order *models.Order) error {
	var (
		err error
	)
	if order.Status == "PROCESSING" {
		_, err = p.db.ExecContext(
			ctx,
			"UPDATE orders SET status = $1 WHERE number = $2;",
			order.Status,
			number,
		)
	} else {
		_, err = p.db.ExecContext(
			ctx,
			"UPDATE orders SET status = $1, processed_at = $2 WHERE number = $3;",
			order.Status,
			order.ProcessedAt.Unix(),
			number,
		)
	}
	if err != nil {
		p.logger.Error().Caller().Msg("unable to update order status")
		return err
	}

	return nil
}

// SelectBalanceByUser calculates amount of points currently
// awailable to user and amount of already withdrawn points.
func (p *postgres) SelectBalanceByUser(ctx context.Context, userID int) (*models.Balance, error) {
	p.logger.Debug().Caller().Msgf("calculating balance for user '%d'", userID)

	b := new(models.Balance)

	rows := p.db.QueryRow(
		"SELECT SUM(amount) FROM posting WHERE user_id = $1;",
		userID,
	)
	var sum sql.NullInt64
	if err := rows.Scan(&sum); err != nil {
		p.logger.Error().Caller().Msg("unable to scan query result for current balance")
		return nil, err
	}
	if rows.Err() != nil {
		p.logger.Error().Caller().Msg("unable to execute query for current balance")
		return nil, rows.Err()
	}

	if !sum.Valid {
		b.Current = 0
	} else {
		b.Current = models.Point(sum.Int64)
	}

	rows = p.db.QueryRow(
		"SELECT ABS(SUM(amount)) FROM posting WHERE user_id = $1 AND journal_id IN (SELECT id FROM balance_journal WHERE type = 'withdrawal');",
		userID,
	)
	if err := rows.Scan(&sum); err != nil {
		p.logger.Error().Caller().Msg("unable to scan query result for withdrawn amount")
		return nil, err
	}
	if rows.Err() != nil {
		p.logger.Error().Caller().Msg("unable to execute query for withdrawn amount")
		return nil, rows.Err()
	}

	if !sum.Valid {
		b.Withdrawn = 0
	} else {
		b.Withdrawn = models.Point(sum.Int64)
	}

	p.logger.Debug().Caller().Msgf("calculated balance for user '%d'", userID)

	return b, nil
}

// InsertWithdrawal insert amount of withdrawn points into
// posting table.
func (p *postgres) InsertWithdrawal(ctx context.Context, userID int, amount float64, orderID int64) error {
	p.logger.Debug().Caller().Msgf("withdrawing %.2f from user '%d'", amount, userID)

	tx, err := p.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: false})
	if err != nil {
		p.logger.Error().Caller().Msg("unable to start a transaction")
		return err
	}
	defer tx.Rollback()

	row := p.db.QueryRowContext(
		ctx,
		"INSERT INTO balance_journal(type) VALUES ('withdrawal') RETURNING id;",
	)

	var id int
	if err := row.Scan(&id); err != nil {
		p.logger.Error().Caller().Msg("unable to scan query result")
		return err
	}
	if row.Err() != nil {
		p.logger.Error().Caller().Msg("unable to execute query for updating balance")
		return err
	}

	_, err = p.db.ExecContext(
		ctx,
		"INSERT INTO posting(user_id, order_id, journal_id, amount) VALUES ($1, $2, $3, $4);",
		userID,
		orderID,
		id,
		-models.ToPoints(amount),
	)
	if err != nil {
		p.logger.Error().Caller().Msg("unable to execute query")
		return err
	}

	row = p.db.QueryRowContext(
		ctx,
		"INSERT INTO balance_journal(type) VALUES ('deposit') RETURNING id;",
	)

	if err := row.Scan(&id); err != nil {
		p.logger.Error().Caller().Msg("unable to scan query result")
		return err
	}
	if row.Err() != nil {
		p.logger.Error().Caller().Msg("unable to execute query for updating balance")
		return err
	}

	_, err = p.db.ExecContext(
		ctx,
		"INSERT INTO posting(user_id, order_id, journal_id, amount) VALUES (1, $1, $2, $3);",
		orderID,
		id,
		models.ToPoints(amount),
	)
	if err != nil {
		p.logger.Error().Caller().Msg("unable to execute query")
		return err
	}

	return tx.Commit()
}

// InsertAccrual insert amount of added points into
// posting table.
func (p *postgres) InsertAccrual(ctx context.Context, userID int, amount float64, orderID int64) error {
	p.logger.Debug().Caller().Msgf("depositing %.2f to user '%d'", amount, userID)

	tx, err := p.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: false})
	if err != nil {
		p.logger.Error().Caller().Msg("unable to start a transaction")
		return err
	}
	defer tx.Rollback()

	row := p.db.QueryRowContext(
		ctx,
		"INSERT INTO balance_journal(type) VALUES ('deposit') RETURNING id;",
	)

	var id int
	if err := row.Scan(&id); err != nil {
		p.logger.Error().Caller().Msg("unable to scan query result")
		return err
	}
	if row.Err() != nil {
		p.logger.Error().Caller().Msg("unable to execute query for updating balance")
		return err
	}

	_, err = p.db.ExecContext(
		ctx,
		"INSERT INTO posting(user_id, order_id, journal_id, amount) VALUES ($1, $2, $3, $4);",
		userID,
		orderID,
		id,
		models.ToPoints(amount),
	)
	if err != nil {
		p.logger.Error().Caller().Msg("unable to execute query")
		return err
	}

	row = p.db.QueryRowContext(
		ctx,
		"INSERT INTO balance_journal(type) VALUES ('withdrawal') RETURNING id;",
	)

	if err := row.Scan(&id); err != nil {
		p.logger.Error().Caller().Msg("unable to scan query result")
		return err
	}
	if row.Err() != nil {
		p.logger.Error().Caller().Msg("unable to execute query for updating balance")
		return err
	}

	_, err = p.db.ExecContext(
		ctx,
		"INSERT INTO posting(user_id, order_id, journal_id, amount) VALUES (1, $1, $2, $3);",
		orderID,
		id,
		-models.ToPoints(amount),
	)
	if err != nil {
		p.logger.Error().Caller().Msg("unable to execute query")
		return err
	}

	return tx.Commit()
}

// SelectWithdrawalsByUser gather order's number, sum and time of processing
// for provided user ID.
func (p *postgres) SelectWithdrawalsByUser(ctx context.Context, userID int) ([]*models.Order, error) {
	p.logger.Debug().Caller().Msgf("selecting withdrawals for user '%d'", userID)

	rows, err := p.db.Query(
		"SELECT o.number, ABS(p.amount), o.processed_at FROM posting AS p INNER JOIN orders AS o ON p.order_id = o.id WHERE p.user_id = $1 AND p.journal_id IN (SELECT id FROM balance_journal WHERE type = 'withdrawal');",
		userID,
	)
	if err != nil {
		p.logger.Error().Caller().Msg("unable to execute query")
		return nil, rows.Err()
	}

	orders := make([]*models.Order, 0)
	for rows.Next() {
		order := new(models.Order)
		var unixProcessed int64
		if err := rows.Scan(&order.Number, &order.Sum, &unixProcessed); err != nil {
			p.logger.Error().Caller().Msg("unable to scan query result")
			return nil, err
		}
		order.ProcessedAt = time.Unix(unixProcessed, 0)
		orders = append(orders, order)
	}

	if rows.Err() != nil {
		p.logger.Error().Caller().Msg("unable to execute query")
		return nil, rows.Err()
	}

	p.logger.Debug().Caller().Msgf("found %d withdrawals for user '%d'", len(orders), userID)

	return orders, nil
}
