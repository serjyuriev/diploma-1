package handlers

import (
	"net/http"

	"github.com/rs/zerolog"
	"github.com/serjyuriev/diploma-1/internal/pkg/repository"
	"github.com/serjyuriev/diploma-1/internal/pkg/service"
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
	repo   repository.Repository
	svc    service.Service
}

func MakeHandlers(logger zerolog.Logger) (Handlers, error) {
	logger.Debug().Msg("preparing handlers")

	repo, err := repository.NewPostgres(logger)
	if err != nil {
		logger.Error().Msg("unable to initialize postgres repository")
		return nil, err
	}

	svc, err := service.NewService(logger, repo)
	if err != nil {
		logger.Error().Msg("unable to initialize new service")
		return nil, err
	}

	return &handlers{
		logger: logger,
		repo:   repo,
		svc:    svc,
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
