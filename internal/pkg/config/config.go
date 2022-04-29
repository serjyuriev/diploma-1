package config

import (
	"log"
	"sync"
	"time"

	"github.com/caarlos0/env"
)

type Config struct {
	RunAddress                 string `env:"RUN_ADDRESS"`
	DatabaseURI                string `env:"DATABASE_URI"`
	AccrualSystemAddress       string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	AccrualSystemPollPeriodInt int    `env:"ACCRUAL_SYSTEM_POLL_PERIOD" envDefault:"2"`
	AccrualSystemPollPeriod    time.Duration
	MigrationsScriptsPath      string `env:"MIGRATION_SCRIPTS_PATH" envDefault:"file://scripts/migrations/"`
	SigningKey                 string `env:"SIGNING_KEY" envDefault:"gopherkey"`
}

var (
	cfg  *Config
	once sync.Once
)

// GetConfig parses flags and environment variables once,
// returning Config struct.
func GetConfig() Config {
	once.Do(func() {
		cfg = &Config{}

		if err := env.Parse(cfg); err != nil {
			log.Printf("unable to load values from environment variables: %v", err)

			// flag.StringVar(&cfg.RunAddress, "a", "", "address and port for starting service on")
			// flag.StringVar(&cfg.DatabaseURI, "d", "", "data source name")
			// flag.StringVar(&cfg.AccrualSystemAddress, "r", "", "address of accrual system")
			// flag.Parse()
		}

		cfg.AccrualSystemPollPeriod = time.Duration(cfg.AccrualSystemPollPeriodInt) * time.Second

		log.Printf("%v\n", cfg)
	})

	return *cfg
}
