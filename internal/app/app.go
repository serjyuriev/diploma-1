package app

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/serjyuriev/diploma-1/internal/pkg/config"
	"github.com/serjyuriev/diploma-1/internal/pkg/handlers"
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
	return nil
}
