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
	"github.com/serjyuriev/diploma-1/internal/pkg/accrual"
	"github.com/serjyuriev/diploma-1/internal/pkg/config"
	"github.com/serjyuriev/diploma-1/internal/pkg/models"
	"github.com/serjyuriev/diploma-1/internal/pkg/repository"
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
	signingKey     []byte
	expireDuration time.Duration
}

// NewService creates new instance of service structure.
func NewService(logger zerolog.Logger, repo repository.Repository) (Service, error) {
	logger.Debug().Caller().Msg("initializing service")
	svc := &service{
		config:         config.GetConfig(),
		accrual:        accrual.NewAccrualClient(logger),
		logger:         logger,
		repo:           repo,
		jobChan:        make(chan *job),
		hashSalt:       "gopher",
		signingKey:     []byte("gopherkey"),
		expireDuration: 10 * time.Minute,
	}
	svc.polling(5)
	return svc, nil
}

// RegisterUser inserts new user credentials into database and
// tries to log user in.
func (svc *service) RegisterUser(ctx context.Context, user *models.User) (string, error) {
	dbUser := &models.User{Login: user.Login}

	pwd := sha1.New()
	pwd.Write([]byte(user.Password))
	pwd.Write([]byte(svc.hashSalt))
	dbUser.Password = fmt.Sprintf("%x", pwd.Sum(nil))

	if err := svc.repo.InsertUser(ctx, dbUser); err != nil {
		svc.logger.Error().Caller().Msg("unable to insert user into database")
		return "", err
	}

	return svc.LoginUser(ctx, user)
}

// LoginUser gathers user credentials from database
// and compares them with provided by caller,
// returning error in case they don't match.
func (svc *service) LoginUser(ctx context.Context, user *models.User) (string, error) {
	svc.logger.Debug().Caller().Msgf("logging in user '%s'", user.Login)

	pwd := sha1.New()
	pwd.Write([]byte(user.Password))
	pwd.Write([]byte(svc.hashSalt))
	user.Password = fmt.Sprintf("%x", pwd.Sum(nil))

	dbUser, err := svc.repo.SelectUser(ctx, user.Login)
	if err != nil {
		svc.logger.Error().Caller().Msg("unable to select user from database")
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

		svc.logger.Debug().Caller().Msgf("user '%s' logged in", user.Login)
		return token.SignedString(svc.signingKey)
	}

	svc.logger.Info().Caller().Msgf("unsuccessful attempt to login for '%s'", user.Login)
	return "", ErrNotRegistered
}

func (svc *service) CreateNewOrder(ctx context.Context, number string, userID int) error {
	svc.logger.Debug().Caller().Msgf("trying to create new order '%s' for user '%d'", number, userID)

	if !svc.validateOrderNumber(number) {
		svc.logger.Warn().Caller().Msgf("order number %s is not valid", number)
		return ErrNotValidOrderNumber
	}

	o, err := svc.repo.SelectOrderByNumber(ctx, number)
	if err == nil {
		if o.UserID == userID {
			return ErrOrderAddedByUser
		} else {
			return ErrOrderAddedByAnotherUser
		}
	}

	if !errors.Is(err, sql.ErrNoRows) {
		svc.logger.Error().Caller().Msg("unable to get order from database")
		return err
	}

	orderID, err := svc.repo.InsertOrder(ctx, number, userID)
	if err != nil {
		svc.logger.Error().Caller().Msg("unable to insert order into database")
		return err
	}

	resc := svc.getOrderAccrualInfo(ctx, number)

	go func() {
		for order := range resc {
			svc.logger.Debug().Caller().Msgf("accrual status = %s", order.AccrualStatus)
			if order.AccrualStatus == "REGISTERED" || order.AccrualStatus == "PROCESSING" {
				if err := svc.repo.UpdateOrderStatus(ctx, number, &models.Order{
					Number: number,
					Status: "PROCESSING",
				}); err != nil {
					svc.logger.Error().Caller().Msg("unable to update order status")
				}
			} else if order.AccrualStatus == "INVALID" {
				if err := svc.repo.UpdateOrderStatus(ctx, number, &models.Order{
					Number:      number,
					Status:      "INVALID",
					ProcessedAt: time.Now(),
				}); err != nil {
					svc.logger.Error().Caller().Msg("unable to update order status")
				}
			} else if order.AccrualStatus == "PROCESSED" {
				if err := svc.repo.InsertAccrual(ctx, userID, order.Accrual.Float64(), orderID); err != nil {
					svc.logger.Error().Caller().Msg("unable to update user balance")
				}
				if err := svc.repo.UpdateOrderStatus(ctx, number, &models.Order{
					Number:      number,
					Status:      "PROCESSED",
					ProcessedAt: time.Now(),
				}); err != nil {
					svc.logger.Error().Caller().Msg("unable to update order status")
				}
			}
		}
	}()

	return nil
}

// WithdrawPoints checks if user has enough points
// and then updates its balance.
func (svc *service) WithdrawPoints(ctx context.Context, userID int, amount float64, order string) error {
	svc.logger.Debug().Caller().Msgf("withdrawing %.2f points for order '%s' of user %d", amount, order, userID)

	if !svc.validateOrderNumber(order) {
		svc.logger.Warn().Caller().Msgf("order number %s is not valid", order)
		return ErrNotValidOrderNumber
	}

	b, err := svc.repo.SelectBalanceByUser(ctx, userID)
	if err != nil {
		svc.logger.Error().Caller().Msg("unable to select user's balance from database")
		return err
	}

	if b.Current.Float64() < amount {
		return ErrNotEnoughPoints
	}

	if err := svc.repo.InsertWithdrawal(ctx, userID, amount); err != nil {
		svc.logger.Error().Caller().Msg("unable to update user's balance in database")
		return err
	}

	return nil
}

// validateOrderNumber checks if provided number is correct
// using Luhn algorithm.
func (svc *service) validateOrderNumber(number string) bool {
	_, err := strconv.Atoi(number)
	if err != nil {
		svc.logger.Error().Caller().Msg("unable to convert order number to integer")
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

	return sum%10 == 0
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
	ticker := time.NewTicker(time.Duration(svc.config.AccrualSystemSurveyPeriod * 1000000000))
	o := new(models.Order)
	var err error

	for range ticker.C {
		o, err = svc.accrual.GetOrderStatus(ctx, order.Number)
		if err != nil {
			svc.logger.Err(err).Msg("unable to get order status from accrual system")
			close(out)
			return
		}

		svc.logger.Info().Caller().Msgf("%v", o)
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
