package handlers

import (
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
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (h *handlers) GetUserOrdersHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
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
	userID := r.Context().Value("user_id").(int)
	b, err := h.repo.SelectBalanceByUser(r.Context(), userID)
	if err != nil {
		h.logger.Err(err).Caller().Msg("unable to get user balance")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	json, err := json.Marshal(b)
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
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (h *handlers) GetUserWithdrawalsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

type getUserOrdersResponse struct {
	Number     string  `json:"number"`
	Status     string  `json:"string"`
	Accrual    float64 `json:"accrual"`
	UploadedAt string  `json:"uploaded_at"`
}
