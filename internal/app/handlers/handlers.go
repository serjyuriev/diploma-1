package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"github.com/serjyuriev/diploma-1/internal/app/repository"
	"github.com/serjyuriev/diploma-1/internal/app/service"
	"github.com/serjyuriev/diploma-1/internal/pkg/models"
)

type ContextKey string

type Handlers interface {
	RegisterUserHandler(w http.ResponseWriter, r *http.Request)
	LoginUserHandler(w http.ResponseWriter, r *http.Request)
	PostUserOrderHandler(w http.ResponseWriter, r *http.Request)
	GetUserOrdersHandler(w http.ResponseWriter, r *http.Request)
	GetUserBalanceHandler(w http.ResponseWriter, r *http.Request)
	WithdrawUserPointsHandler(w http.ResponseWriter, r *http.Request)
	GetUserWithdrawalsHandler(w http.ResponseWriter, r *http.Request)
}

type handlers struct {
	logger zerolog.Logger
	repo   repository.Repository
	svc    service.Service
}

// MakeHandlers creates new instance of Handlers interface.
func MakeHandlers(logger zerolog.Logger) (Handlers, error) {
	logger.Debug().Msg("preparing handlers")

	repo, err := repository.NewPostgres(logger)
	if err != nil {
		logger.Error().Caller().Msg("unable to initialize postgres repository")
		return nil, err
	}

	svc, err := service.NewService(logger, repo)
	if err != nil {
		logger.Error().Caller().Msg("unable to create new service")
		return nil, err
	}

	return &handlers{
		logger: logger,
		repo:   repo,
		svc:    svc,
	}, nil
}

func (h *handlers) RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	h.logger.Info().Msg("POST /api/user/register")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.
			Err(err).
			Caller().
			Int("code", http.StatusBadRequest).
			Msg("unable to read request body")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	user := new(models.User)
	if err := json.Unmarshal(body, user); err != nil {
		h.logger.
			Err(err).
			Caller().
			Int("code", http.StatusBadRequest).
			Msg("unable to map json to user")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	token, err := h.svc.RegisterUser(r.Context(), user)
	if err != nil {
		if errors.Is(err, service.ErrNotRegistered) {
			h.logger.
				Err(err).
				Caller().
				Str("user", user.Login).
				Int("code", http.StatusInternalServerError).
				Msg("unable to login just registered user")
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		h.logger.
			Err(err).
			Caller().
			Str("user", user.Login).
			Int("code", http.StatusInternalServerError).
			Msg("unable to register user")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Authorization", fmt.Sprintf("Bearer %s", token))
	w.WriteHeader(http.StatusOK)
	h.logger.
		Info().
		Str("user", user.Login).
		Int("code", http.StatusOK).
		Msg("POST /api/user/register")
}

func (h *handlers) LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	h.logger.Info().Msg("POST /api/user/login")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.
			Err(err).
			Caller().
			Int("code", http.StatusBadRequest).
			Msg("unable to read request body")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	user := new(models.User)
	if err := json.Unmarshal(body, user); err != nil {
		h.logger.
			Err(err).
			Caller().
			Int("code", http.StatusBadRequest).
			Msg("unable to map json to user")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	token, err := h.svc.LoginUser(r.Context(), user)
	if err != nil {
		if errors.Is(err, service.ErrNotRegistered) {
			h.logger.
				Info().
				Str("user", user.Login).
				Int("code", http.StatusUnauthorized).
				Msg("user is not registered")
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		h.logger.
			Err(err).
			Caller().
			Str("user", user.Login).
			Int("code", http.StatusInternalServerError).
			Msg("unexpected error occured while trying to log user in")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Authorization", fmt.Sprintf("Bearer %s", token))
	w.WriteHeader(http.StatusOK)
	h.logger.
		Info().
		Str("user", user.Login).
		Int("code", http.StatusOK).
		Msg("POST /api/user/login")
}

func (h *handlers) PostUserOrderHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(ContextKey("user_id")).(int)
	h.logger.Info().Int("user_id", userID).Msg("POST /api/user/orders")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.
			Err(err).
			Caller().
			Int("user_id", userID).
			Int("code", http.StatusBadRequest).
			Msg("unable to read request body")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := h.svc.CreateNewOrder(context.Background(), string(body), userID); err != nil {
		if errors.Is(err, service.ErrNotValidOrderNumber) {
			h.logger.
				Info().
				Str("order_number", string(body)).
				Int("user_id", userID).
				Int("code", http.StatusUnprocessableEntity).
				Msg("order number is not valid")
			w.WriteHeader(http.StatusUnprocessableEntity)
		} else if errors.Is(err, service.ErrOrderAddedByUser) {
			h.logger.
				Info().
				Str("order_number", string(body)).
				Int("user_id", userID).
				Int("code", http.StatusOK).
				Msg("order already added by user")
			w.WriteHeader(http.StatusOK)
		} else if errors.Is(err, service.ErrOrderAddedByAnotherUser) {
			h.logger.
				Info().
				Str("order_number", string(body)).
				Int("user_id", userID).
				Int("code", http.StatusConflict).
				Msg("order already added by another user")
			w.WriteHeader(http.StatusConflict)
		} else {
			h.logger.
				Error().
				Caller().
				Str("order_number", string(body)).
				Int("user_id", userID).
				Int("code", http.StatusInternalServerError).
				Msg("unable to create new order")
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusAccepted)
	h.logger.
		Info().
		Str("order_number", string(body)).
		Int("user_id", userID).
		Int("code", http.StatusAccepted).
		Msg("POST /api/user/orders")
}

func (h *handlers) GetUserOrdersHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(ContextKey("user_id")).(int)
	h.logger.Info().Int("user_id", userID).Msg("GET /api/user/orders")
	orders, err := h.repo.SelectOrdersByUser(r.Context(), userID)
	if err != nil {
		h.logger.
			Err(err).
			Caller().
			Int("user_id", userID).
			Int("code", http.StatusInternalServerError).
			Msg("unable to select orders from database")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		h.logger.
			Info().
			Int("user_id", userID).
			Int("code", http.StatusNoContent).
			Msg("there is no user orders in system")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	res := make([]getUserOrdersResponse, 0)
	for _, order := range orders {
		sr := getUserOrdersResponse{
			Number:     order.Number,
			Status:     order.Status,
			Accrual:    order.Accrual.Float64(),
			UploadedAt: order.UploadedAt.Format(time.RFC3339),
		}
		res = append(res, sr)
	}

	json, err := json.Marshal(res)
	if err != nil {
		h.logger.
			Err(err).
			Caller().
			Int("user_id", userID).
			Int("code", http.StatusInternalServerError).
			Msg("unable to marshal response")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
	h.logger.
		Info().
		Int("user_id", userID).
		Int("code", http.StatusOK).
		Msg("GET /api/user/orders")
}

func (h *handlers) GetUserBalanceHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(ContextKey("user_id")).(int)
	h.logger.Info().Int("user_id", userID).Msg("GET /api/user/balance")
	b, err := h.repo.SelectBalanceByUser(r.Context(), userID)
	if err != nil {
		h.logger.
			Err(err).
			Caller().
			Int("user_id", userID).
			Int("code", http.StatusInternalServerError).
			Msg("unable to select balance from database")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	res := getUserBalanceResponse{
		Current:   b.Current.Float64(),
		Withdrawn: b.Withdrawn.Float64(),
	}

	json, err := json.Marshal(res)
	if err != nil {
		h.logger.
			Err(err).
			Caller().
			Int("user_id", userID).
			Int("code", http.StatusInternalServerError).
			Msg("unable to marshal response")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
	h.logger.
		Info().
		Int("user_id", userID).
		Int("code", http.StatusOK).
		Msg("GET /api/user/balance")
}

func (h *handlers) WithdrawUserPointsHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(ContextKey("user_id")).(int)
	h.logger.Info().Int("user_id", userID).Msg("POST /api/user/balance/withdraw")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.
			Err(err).
			Caller().
			Int("user_id", userID).
			Int("code", http.StatusBadRequest).
			Msg("unable to read request body")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	req := new(withdrawUserPointsRequest)
	if err := json.Unmarshal(body, req); err != nil {
		h.logger.
			Err(err).
			Caller().
			Int("user_id", userID).
			Int("code", http.StatusBadRequest).
			Msg("unable to unmarshal request body")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if err := h.svc.WithdrawPoints(r.Context(), userID, req.Sum, req.OrderNumber); err != nil {
		if errors.Is(err, service.ErrNotEnoughPoints) {
			h.logger.
				Info().
				Str("order_number", req.OrderNumber).
				Int("user_id", userID).
				Int("code", http.StatusPaymentRequired).
				Msg("user does not have enough points")
			w.WriteHeader(http.StatusPaymentRequired)
		} else if errors.Is(err, service.ErrNotValidOrderNumber) {
			h.logger.
				Info().
				Str("order_number", req.OrderNumber).
				Int("user_id", userID).
				Int("code", http.StatusUnprocessableEntity).
				Msg("order number is not valid")
			w.WriteHeader(http.StatusUnprocessableEntity)
		} else {
			h.logger.
				Err(err).
				Caller().
				Str("order_number", req.OrderNumber).
				Int("user_id", userID).
				Int("code", http.StatusInternalServerError).
				Msg("unable to withdraw points")
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	h.logger.
		Info().
		Str("order_number", req.OrderNumber).
		Int("user_id", userID).
		Int("code", http.StatusOK).
		Msg("POST /api/user/balance/withdraw")
}

func (h *handlers) GetUserWithdrawalsHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(ContextKey("user_id")).(int)
	h.logger.Info().Int("user_id", userID).Msg("GET /api/user/balance/withdrawals")
	withdrawals, err := h.repo.SelectWithdrawalsByUser(r.Context(), userID)
	if err != nil {
		h.logger.
			Err(err).
			Caller().
			Int("user_id", userID).
			Int("code", http.StatusInternalServerError).
			Msg("unable to get user withdrawals")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		h.logger.
			Info().
			Int("user_id", userID).
			Int("code", http.StatusNoContent).
			Msg("user does not have any withdrawals")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	res := make([]getUserWithdrawalsResponse, 0)
	for _, withdrawal := range withdrawals {
		sr := getUserWithdrawalsResponse{
			Number:      withdrawal.Number,
			Sum:         withdrawal.Sum.Float64(),
			ProcessedAt: withdrawal.ProcessedAt.Format(time.RFC3339),
		}
		res = append(res, sr)
	}

	json, err := json.Marshal(res)
	if err != nil {
		h.logger.
			Err(err).
			Caller().
			Int("user_id", userID).
			Int("code", http.StatusInternalServerError).
			Msg("unable to marshal response")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(json)
	h.logger.
		Info().
		Int("user_id", userID).
		Int("code", http.StatusOK).
		Msg("GET /api/user/balance/withdrawals")
}

type getUserBalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type getUserOrdersResponse struct {
	Number     string  `json:"number"`
	Status     string  `json:"status"`
	Accrual    float64 `json:"accrual,omitempty"`
	UploadedAt string  `json:"uploaded_at"`
}

type getUserWithdrawalsResponse struct {
	Number      string  `json:"number"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

type withdrawUserPointsRequest struct {
	OrderNumber string  `json:"order"`
	Sum         float64 `json:"sum"`
}
