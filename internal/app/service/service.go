package service

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/rs/zerolog"
	"github.com/serjyuriev/diploma-1/internal/app/repository"
	"github.com/serjyuriev/diploma-1/internal/pkg/accrual"
	"github.com/serjyuriev/diploma-1/internal/pkg/config"
	"github.com/serjyuriev/diploma-1/internal/pkg/models"
)

var (
	errNotImplemented = errors.New("method not implemented yet")

	ErrNotEnoughPoints         = errors.New("not enough points")
	ErrNotRegistered           = errors.New("user with provided credentials isn't registered")
	ErrNotValidOrderNumber     = errors.New("order number is not valid")
	ErrOrderAddedByUser        = errors.New("order already added by current user")
	ErrOrderAddedByAnotherUser = errors.New("order already added by another user")
)

type job struct {
	ctx     context.Context
	order   *models.Order
	resChan chan *models.Order
}

type Service interface {
	RegisterUser(ctx context.Context, user *models.User) (string, error)
	LoginUser(ctx context.Context, user *models.User) (string, error)
	CreateNewOrder(ctx context.Context, number string, userID int) error
	WithdrawPoints(ctx context.Context, userID int, amount float64, order string) error
}

type service struct {
	config         config.Config
	accrual        accrual.Accrual
	logger         zerolog.Logger
	repo           repository.Repository
	jobChan        chan *job
	hashSalt       string
	expireDuration time.Duration
}

// NewService creates new instance of service structure.
func NewService(logger zerolog.Logger, repo repository.Repository) (Service, error) {
	logger.Debug().Msg("initializing service")
	svc := &service{
		config:         config.GetConfig(),
		accrual:        accrual.NewAccrualClient(logger),
		logger:         logger,
		repo:           repo,
		jobChan:        make(chan *job),
		hashSalt:       "gopher",
		expireDuration: 10 * time.Minute,
	}
	svc.polling(5)
	return svc, nil
}

// RegisterUser inserts new user credentials into database and
// tries to log user in.
func (svc *service) RegisterUser(ctx context.Context, user *models.User) (string, error) {
	svc.logger.Debug().Str("user", user.Login).Msg("registering user")
	dbUser := &models.User{Login: user.Login}

	pwd := sha1.New()
	pwd.Write([]byte(user.Password))
	pwd.Write([]byte(svc.hashSalt))
	dbUser.Password = fmt.Sprintf("%x", pwd.Sum(nil))

	if err := svc.repo.InsertUser(ctx, dbUser); err != nil {
		svc.logger.Error().Caller().Str("user", user.Login).Msg("unable to insert user into database")
		return "", err
	}

	svc.logger.Debug().Str("user", user.Login).Msg("user registered successfully")
	return svc.LoginUser(ctx, user)
}

// LoginUser gathers user credentials from database
// and compares them with provided by caller,
// returning error in case they don't match.
func (svc *service) LoginUser(ctx context.Context, user *models.User) (string, error) {
	svc.logger.Debug().Str("user", user.Login).Msg("logging in user")

	pwd := sha1.New()
	pwd.Write([]byte(user.Password))
	pwd.Write([]byte(svc.hashSalt))
	user.Password = fmt.Sprintf("%x", pwd.Sum(nil))

	dbUser, err := svc.repo.SelectUser(ctx, user.Login)
	if err != nil {
		svc.logger.Error().Caller().Str("user", user.Login).Msg("unable to select user from database")
		return "", err
	}

	if dbUser.Login == user.Login && dbUser.Password == user.Password {
		token := jwt.NewWithClaims(
			jwt.SigningMethodHS256,
			models.Claims{
				StandardClaims: jwt.StandardClaims{
					ExpiresAt: time.Now().Add(svc.expireDuration).Unix(),
					IssuedAt:  time.Now().Unix(),
				},
				UserID: dbUser.ID,
			},
		)

		svc.logger.Debug().Str("user", user.Login).Msg("user logged in")
		return token.SignedString([]byte(svc.config.SigningKey))
	}

	svc.logger.Info().Str("user", user.Login).Msg("unsuccessful attempt to login")
	return "", ErrNotRegistered
}

// CreateNewOrder validates order number using Luhn algorithm,
// checks whether order with such number already exists
// and if not creates new order.
func (svc *service) CreateNewOrder(ctx context.Context, number string, userID int) error {
	svc.logger.
		Debug().
		Str("order_number", number).
		Int("user_id", userID).
		Msg("trying to create new order")

	if !svc.validateOrderNumber(number) {
		svc.logger.
			Debug().
			Str("order_number", number).
			Int("user_id", userID).
			Msg("order number is not valid")
		return ErrNotValidOrderNumber
	}

	o, err := svc.repo.SelectOrderByNumber(ctx, number)
	if err == nil {
		svc.logger.
			Debug().
			Str("order_number", number).
			Int("user_id", userID).
			Msg("order was found in database")
		if o.UserID == userID {
			return ErrOrderAddedByUser
		} else {
			return ErrOrderAddedByAnotherUser
		}
	}

	if !errors.Is(err, sql.ErrNoRows) {
		svc.logger.
			Error().
			Caller().
			Str("order_number", number).
			Int("user_id", userID).
			Msg("unable to select order from database")
		return err
	}

	orderID, err := svc.repo.InsertOrder(ctx, number, userID)
	if err != nil {
		svc.logger.
			Error().
			Caller().
			Str("order_number", number).
			Int("user_id", userID).
			Msg("unable to insert order into database")
		return err
	}

	resc := svc.getOrderAccrualInfo(ctx, number)

	go func() {
		for order := range resc {
			svc.logger.
				Debug().
				Str("order_number", number).
				Int("user_id", userID).
				Msgf("accrual status = %s", order.AccrualStatus)
			if order.AccrualStatus == "REGISTERED" || order.AccrualStatus == "PROCESSING" {
				if err := svc.repo.UpdateOrderStatus(ctx, number, &models.Order{
					Number: number,
					Status: "PROCESSING",
				}); err != nil {
					svc.logger.
						Error().
						Caller().
						Str("order_number", number).
						Int("user_id", userID).
						Msg("unable to update order status")
				}
			} else if order.AccrualStatus == "INVALID" {
				if err := svc.repo.UpdateOrderStatus(ctx, number, &models.Order{
					Number:      number,
					Status:      "INVALID",
					ProcessedAt: time.Now(),
				}); err != nil {
					svc.logger.
						Error().
						Caller().
						Str("order_number", number).
						Int("user_id", userID).
						Msg("unable to update order status")
				}
			} else if order.AccrualStatus == "PROCESSED" {
				if err := svc.repo.InsertAccrual(ctx, userID, order.Accrual.Float64(), orderID); err != nil {
					svc.logger.
						Error().
						Caller().
						Str("order_number", number).
						Int("user_id", userID).
						Msg("unable to update user balance")
				}
				if err := svc.repo.UpdateOrderStatus(ctx, number, &models.Order{
					Number:      number,
					Status:      "PROCESSED",
					ProcessedAt: time.Now(),
				}); err != nil {
					svc.logger.
						Error().
						Caller().
						Str("order_number", number).
						Int("user_id", userID).
						Msg("unable to update order status")
				}
			}
		}
	}()

	svc.logger.
		Debug().
		Str("order_number", number).
		Int("user_id", userID).
		Msg("order is being processed")
	return nil
}

// WithdrawPoints checks if user has enough points
// and then updates its balance.
func (svc *service) WithdrawPoints(ctx context.Context, userID int, amount float64, order string) error {
	svc.logger.
		Debug().
		Str("order_number", order).
		Int("user_id", userID).
		Msgf("withdrawing %.2f points", amount)

	if !svc.validateOrderNumber(order) {
		svc.logger.
			Debug().
			Str("order_number", order).
			Int("user_id", userID).
			Msg("order number is not valid")
		return ErrNotValidOrderNumber
	}

	b, err := svc.repo.SelectBalanceByUser(ctx, userID)
	if err != nil {
		svc.logger.
			Error().
			Caller().
			Str("order_number", order).
			Int("user_id", userID).
			Msg("unable to select user balance from database")
		return err
	}

	if b.Current.Float64() < amount {
		svc.logger.
			Debug().
			Str("order_number", order).
			Int("user_id", userID).
			Msg("user does not have enough points")
		return ErrNotEnoughPoints
	}

	if err := svc.repo.InsertWithdrawal(ctx, userID, amount); err != nil {
		svc.logger.
			Error().
			Caller().
			Str("order_number", order).
			Int("user_id", userID).
			Msg("unable to update user balance in database")
		return err
	}

	svc.logger.
		Debug().
		Str("order_number", order).
		Int("user_id", userID).
		Msg("points were successfully withdrawn")
	return nil
}

// validateOrderNumber checks if provided number is correct
// using Luhn algorithm.
func (svc *service) validateOrderNumber(number string) bool {
	svc.logger.Debug().Str("order_number", number).Msg("validating order number")
	_, err := strconv.Atoi(number)
	if err != nil {
		svc.logger.
			Error().
			Caller().
			Str("order_number", number).
			Bool("is_valid", false).
			Msg("unable to convert order number to integer")
		return false
	}

	sum := 0
	for i := len(number) - 1; i >= 0; i -= 2 {
		d, _ := strconv.Atoi(string(number[i]))
		sum += d
	}

	for i := len(number) - 2; i >= 0; i -= 2 {
		d, _ := strconv.Atoi(string(number[i]))
		d *= 2
		if d > 9 {
			d -= 9
		}
		sum += d
	}

	isValid := sum%10 == 0
	svc.logger.
		Debug().
		Str("order_number", number).
		Bool("is_valid", isValid).
		Msg("order number was validated")
	return isValid
}

func (svc *service) getOrderAccrualInfo(ctx context.Context, number string) chan *models.Order {
	res := make(chan *models.Order)
	svc.jobChan <- &job{
		ctx:     ctx,
		order:   &models.Order{Number: number},
		resChan: res,
	}
	return res
}

func (svc *service) pollAccrualSystem(ctx context.Context, order *models.Order, out chan *models.Order) {
	ticker := time.NewTicker(svc.config.AccrualSystemPollPeriod)

	for range ticker.C {
		o, err := svc.accrual.GetOrderStatus(ctx, order.Number)
		if err != nil {
			svc.logger.
				Err(err).
				Caller().
				Str("order_number", order.Number).
				Msg("unable to get order status from accrual system")
			close(out)
			return
		}

		out <- o

		if o.AccrualStatus == "INVALID" || o.AccrualStatus == "PROCESSED" {
			close(out)
			return
		}
	}
}

func (svc *service) polling(workers int) {
	for i := 0; i < workers; i++ {
		go func() {
			for job := range svc.jobChan {
				svc.pollAccrualSystem(job.ctx, job.order, job.resChan)
			}
		}()
	}
}
