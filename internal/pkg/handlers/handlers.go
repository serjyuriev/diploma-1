package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"github.com/serjyuriev/diploma-1/internal/pkg/models"
	"github.com/serjyuriev/diploma-1/internal/pkg/repository"
	"github.com/serjyuriev/diploma-1/internal/pkg/service"
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
	logger.Debug().Caller().Msg("preparing handlers")

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
	h.logger.Info().Caller().Msg("POST /api/user/register")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Err(err).Caller().Msg("unable to read request body")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	user := new(models.User)
	if err := json.Unmarshal(body, user); err != nil {
		h.logger.Err(err).Caller().Msg("unable to map json to user")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	token, err := h.svc.RegisterUser(r.Context(), user)
	if err != nil {
		if errors.Is(err, service.ErrNotRegistered) {
			h.logger.Err(err).Caller().Msg("unable to login just registered user")
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		h.logger.Err(err).Caller().Msg("unable to register user")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Authorization", fmt.Sprintf("Bearer %s", token))
	w.WriteHeader(http.StatusOK)
}

func (h *handlers) LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	h.logger.Info().Caller().Msg("POST /api/user/login")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Err(err).Caller().Msg("unable to read request body")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	user := new(models.User)
	if err := json.Unmarshal(body, user); err != nil {
		h.logger.Err(err).Caller().Msg("unable to map json to user")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	token, err := h.svc.LoginUser(r.Context(), user)
	if err != nil {
		if errors.Is(err, service.ErrNotRegistered) {
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		h.logger.Err(err).Caller().Msg("unable to login user")
		http.Error(w, "internal server error", http.StatusUnauthorized)
		return
	}

	w.Header().Add("Authorization", fmt.Sprintf("Bearer %s", token))
	w.WriteHeader(http.StatusOK)
}

func (h *handlers) PostUserOrderHandler(w http.ResponseWriter, r *http.Request) {
	h.logger.Info().Caller().Msg("POST /api/user/orders")
	userID := r.Context().Value(ContextKey("user_id")).(int)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Err(err).Caller().Msg("unable to read request body")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := h.svc.CreateNewOrder(context.Background(), string(body), userID); err != nil {
		if errors.Is(err, service.ErrNotValidOrderNumber) {
			w.WriteHeader(http.StatusUnprocessableEntity)
		} else if errors.Is(err, service.ErrOrderAddedByUser) {
			w.WriteHeader(http.StatusOK)
		} else if errors.Is(err, service.ErrOrderAddedByAnotherUser) {
			w.WriteHeader(http.StatusConflict)
		} else {
			h.logger.Error().Caller().Msg("unable to create new order")
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *handlers) GetUserOrdersHandler(w http.ResponseWriter, r *http.Request) {
	h.logger.Info().Caller().Msg("GET /api/user/orders")
	userID := r.Context().Value(ContextKey("user_id")).(int)
	orders, err := h.repo.SelectOrdersByUser(r.Context(), userID)
	if err != nil {
		h.logger.Err(err).Caller().Msg("unable to get user orders")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
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
		h.logger.Err(err).Caller().Msg("unable to marshal response")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	w.Write(json)
}

func (h *handlers) GetUserBalanceHandler(w http.ResponseWriter, r *http.Request) {
	h.logger.Info().Caller().Msg("GET /api/user/balance")
	userID := r.Context().Value(ContextKey("user_id")).(int)
	b, err := h.repo.SelectBalanceByUser(r.Context(), userID)
	if err != nil {
		h.logger.Err(err).Caller().Msg("unable to get user balance")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	res := getUserBalanceResponse{
		Current:   b.Current.Float64(),
		Withdrawn: b.Withdrawn.Float64(),
	}

	json, err := json.Marshal(res)
	if err != nil {
		h.logger.Err(err).Caller().Msg("unable to marshal response")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	w.Write(json)
}

func (h *handlers) WithdrawUserPointsHandler(w http.ResponseWriter, r *http.Request) {
	h.logger.Info().Caller().Msg("POST /api/user/balance/withdraw")
	userID := r.Context().Value(ContextKey("user_id")).(int)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Err(err).Caller().Msg("unable to read request body")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	req := new(withdrawUserPointsRequest)
	if err := json.Unmarshal(body, req); err != nil {
		h.logger.Err(err).Caller().Msg("unable to unmarshal request body")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if err := h.svc.WithdrawPoints(r.Context(), userID, req.Sum, req.OrderNumber); err != nil {
		if errors.Is(err, service.ErrNotEnoughPoints) {
			w.WriteHeader(http.StatusPaymentRequired)
			return
		} else if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		} else {
			h.logger.Err(err).Caller().Msg("unable to withdraw points")
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
}

func (h *handlers) GetUserWithdrawalsHandler(w http.ResponseWriter, r *http.Request) {
	h.logger.Info().Caller().Msg("GET /api/user/balance/withdrawals")
	userID := r.Context().Value(ContextKey("user_id")).(int)
	withdrawals, err := h.repo.SelectWithdrawalsByUser(r.Context(), userID)
	if err != nil {
		h.logger.Err(err).Caller().Msg("unable to get user withdrawals")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
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
		h.logger.Err(err).Caller().Msg("unable to marshal response")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	w.Write(json)
}

type getUserBalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type getUserOrdersResponse struct {
	Number     string  `json:"number"`
	Status     string  `json:"status"`
	Accrual    float64 `json:"accrual"`
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
