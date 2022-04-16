package handlers

import (
	"net/http"

	"github.com/rs/zerolog"
)

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
}

func MakeHandlers(logger zerolog.Logger) (Handlers, error) {
	return &handlers{
		logger: logger,
	}, nil
}

func (h *handlers) RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (h *handlers) LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (h *handlers) PostUserOrderHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (h *handlers) GetUserOrdersHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (h *handlers) GetUserBalanceHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (h *handlers) WithdrawUserPointsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (h *handlers) GetUserWithdrawalsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
