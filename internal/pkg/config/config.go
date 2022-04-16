package config

import (
	"flag"
	"log"
	"sync"

	"github.com/caarlos0/env"
)

type Config struct {
	RunAddress                string `env:"RUN_ADDRESS"`
	DatabaseUri               string `env:"DATABASE_URI"`
	AccrualSystemAddress      string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	AccrualSystemSurveyPeriod int    `env:"ACCRUAL_SYSTEM_SURVEY_PERIOD" envDefault:"5"`
}

var (
	cfg  Config
	once sync.Once
)

// GetConfig parses flags and environment variables once,
// returning Config struct.
func GetConfig() Config {
	once.Do(func() {
		cfg = Config{}
		flag.StringVar(&cfg.RunAddress, "a", "localhost:8080", "address and port for starting service on")
		flag.StringVar(&cfg.DatabaseUri, "d", "", "data source name")
		flag.StringVar(&cfg.AccrualSystemAddress, "r", "", "address of accrual system")
		flag.Parse()

		if err := env.Parse(cfg); err != nil {
			log.Fatalf("unable to load values from environment variables: %v", err)
		}
	})

	return cfg
}
