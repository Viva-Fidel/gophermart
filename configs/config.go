package config

import (
	"github.com/caarlos0/env/v11"
)

type Config struct {
	Server     ServerConfig
	Db         DbConfig
	Accural    AccuralConfig
	Auth       AuthConfig
}


type ServerConfig struct {
	Address *string `env:"RUN_ADDRESS"`
}

type DbConfig struct {
	DatabaseURI       *string `env:"DATABASE_URI"`
}

type AccuralConfig struct {
	AccuralSystemAdress *string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

type AuthConfig struct {
	JWTSecret string `env:"JWT_SECRET" envDefault:"gophermart-dev-secret"`
	TokenExp string `env:"TOKEN_EXP" envDefault:"24h"`
}

// Загружает конфигурацию
func LoadConfig() (*Config, error) {
	cfg := Config{}

	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
