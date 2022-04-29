package accrual

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog"
	"github.com/serjyuriev/diploma-1/internal/pkg/config"
	"github.com/serjyuriev/diploma-1/internal/pkg/models"
)

var (
	errAccrualTooManyRequests = errors.New("too many requests to accrual system")
	errAccrualInternal        = errors.New("accrual system internal error occured")
)

type Accrual interface {
	GetOrderStatus(ctx context.Context, order string) (*models.Order, error)
}

type accrual struct {
	logger    zerolog.Logger
	systemURL string
}

func NewAccrualClient(logger zerolog.Logger) Accrual {
	return &accrual{
		logger:    logger,
		systemURL: config.GetConfig().AccrualSystemAddress,
	}
}

func (a *accrual) GetOrderStatus(ctx context.Context, order string) (*models.Order, error) {
	a.logger.Debug().Str("order_number", order).Msg("polling accrual system")
	client := http.Client{}
	accrualRequestURL := fmt.Sprintf("%s/api/orders/%s", a.systemURL, order)

	req, err := http.NewRequest(
		http.MethodGet,
		accrualRequestURL,
		nil,
	)
	if err != nil {
		a.logger.
			Error().
			Caller().
			Str("order_number", order).
			Msg("unable to create new request")
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		a.logger.
			Error().
			Caller().
			Str("order_number", order).
			Msg("unable to send request to accrual system")
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusTooManyRequests {
		a.logger.
			Error().
			Caller().
			Str("order_number", order).
			Int("accrual_code", http.StatusTooManyRequests).
			Msg("too many requests to accrual system")
		return nil, errAccrualTooManyRequests
	} else if res.StatusCode == http.StatusInternalServerError {
		a.logger.
			Error().
			Caller().
			Str("order_number", order).
			Int("accrual_code", http.StatusInternalServerError).
			Msg("accrual system internal error occured")
		return nil, errAccrualInternal
	} else {
		a.logger.
			Debug().
			Str("order_number", order).
			Int("accrual_code", res.StatusCode).
			Msg("accrual system responded")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		a.logger.
			Error().
			Caller().
			Str("order_number", order).
			Msg("unable to read response body")
		return nil, err
	}

	accrualResponse := new(accrualResponse)
	if err = json.Unmarshal(body, &accrualResponse); err != nil {
		a.logger.
			Error().
			Caller().
			Str("order_number", order).
			Msg("unable to unmarshal accrual system response")
		return nil, err
	}

	a.logger.Debug().Str("order_number", order).Msg("accrual system was polled")
	return &models.Order{
		Number:        accrualResponse.Order,
		AccrualStatus: accrualResponse.Status,
		Accrual:       models.ToPoints(accrualResponse.Accrual),
	}, nil
}

type accrualResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}
