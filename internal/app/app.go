package app

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/serjyuriev/diploma-1/internal/pkg/config"
)

type App interface {
	Start() error
}

type app struct {
	logger zerolog.Logger
	cfg    config.Config
}

func NewApp() (App, error) {
	return &app{
		logger: zerolog.New(os.Stdout),
		cfg:    config.GetConfig(),
	}, nil
}

func (app *app) Start() error {
	return nil
}
