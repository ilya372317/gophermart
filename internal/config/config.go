package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env"
)

const defaultExpTokenTimeInHours = 12
const defaultMaxOrderInProcessing = 30
const defaultMaxAccrualRequestAttempts = 5
const defaultDelayBetweenRequestsToAccrual = 5

type GophermartConfig struct {
	Host                          string `env:"RUN_ADDRESS"`
	DatabaseDSN                   string `env:"DATABASE_URI"`
	SecretKey                     string `env:"SECRET_KEY"`
	AccrualAddress                string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	ExpTime                       uint   `env:"EXP_TIME"`
	MaxOrderInProcessing          int    `env:"MAX_ORDER_IN_PROCESSING"`
	MaxAccrualRequestAttempts     int    `env:"MAX_ACCRUAL_REQUEST_ATTEMPTS"`
	DelayBetweenRequestsToAccrual int    `env:"DELAY_BETWEEN_REQUEST_TO_ACCRUAL"`
}

func New() (*GophermartConfig, error) {
	cnfg := &GophermartConfig{}
	cnfg.parseFlags()
	err := env.Parse(cnfg)
	if err != nil {
		return nil, fmt.Errorf("failed parse enviroment virables: %w", err)
	}
	return cnfg, nil
}

func (c *GophermartConfig) parseFlags() {
	flag.StringVar(&c.Host, "a", ":8080", "address where server will listen requests")
	flag.StringVar(&c.DatabaseDSN, "d", "", "database DSN string")
	flag.StringVar(&c.SecretKey, "k", "this-need-to-be-replace", "secret key for authentication")
	flag.UintVar(&c.ExpTime, "e", uint(defaultExpTokenTimeInHours), "token expiration time in hours")
	flag.StringVar(&c.AccrualAddress, "r", ":8090", "address of accrual system")
	flag.IntVar(&c.MaxOrderInProcessing, "o", defaultMaxOrderInProcessing,
		"max order wich can be in processing simultaneously")
	flag.IntVar(&c.MaxAccrualRequestAttempts, "ar", defaultMaxAccrualRequestAttempts,
		"count of attempts when system will try get data from accrual")
	flag.IntVar(&c.DelayBetweenRequestsToAccrual, "ad", defaultDelayBetweenRequestsToAccrual,
		"delay between request to accrual in seconds")
	flag.Parse()
}
