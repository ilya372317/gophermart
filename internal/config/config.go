package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env"
)

type GophermartConfig struct {
	Host        string `env:"ADDRESS"`
	DatabaseDSN string `env:"DATABASE_DSN"`
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
	flag.StringVar(&c.DatabaseDSN, "d", "", "Database DSN string")
	flag.Parse()
}
