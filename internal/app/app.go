package app

import (
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
	"github.com/serjyuriev/diploma-1/internal/pkg/config"
	"github.com/serjyuriev/diploma-1/internal/pkg/handlers"
	"github.com/serjyuriev/diploma-1/internal/pkg/middleware"
)

type App interface {
	Start() error
}

type app struct {
	cfg      config.Config
	handlers handlers.Handlers
	logger   zerolog.Logger
}

func NewApp() (App, error) {
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "02-01-2006 15:04:05 MST",
	}
	logger := zerolog.New(output).With().Timestamp().Logger()
	handlers, err := handlers.MakeHandlers(logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("unable to make handlers")
	}

	return &app{
		cfg:      config.GetConfig(),
		handlers: handlers,
		logger:   logger,
	}, nil
}

func (app *app) Start() error {
	r := chi.NewRouter()
	r.Use(middleware.Auth)
	r.Post("/api/user/register", app.handlers.RegisterUserHandler)
	r.Post("/api/user/login", app.handlers.LoginUserHandler)
	r.Post("/api/user/orders", app.handlers.PostUserOrderHandler)
	r.Get("/api/user/orders", app.handlers.GetUserOrdersHandler)
	r.Get("/api/user/balance", app.handlers.GetUserBalanceHandler)
	r.Post("/api/user/balance/withdraw", app.handlers.WithdrawUserPointsHandler)
	r.Get("/api/user/balance/withdrawals", app.handlers.GetUserWithdrawalsHandler)

	app.logger.Info().Msgf("starting application on %s", app.cfg.RunAddress)
	return http.ListenAndServe(app.cfg.RunAddress, r)
}
