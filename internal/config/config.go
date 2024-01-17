package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env"
)

const defaultExpTokenTimeInHours = 12

type GophermartConfig struct {
	Host           string `env:"ADDRESS"`
	DatabaseDSN    string `env:"DATABASE_DSN"`
	SecretKey      string `env:"SECRET_KEY"`
	AccrualAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	ExpTime        uint   `env:"EXP_TIME"`
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
	flag.Parse()
}
