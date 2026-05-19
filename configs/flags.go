package config

import (
	"errors"
	"flag"
	"os"
)

type Flags struct {
	RunAddress          string
	DatabaseURI         string
	AccrualSystemAdress string
	JWTSecret           string
	TokenExp            string
}

func parseFlags(conf *Config, fs *flag.FlagSet, args []string) (*Flags, error) {
	// Инициализирует флаги и берём данные из конфигурации
	flags := &Flags{
		RunAddress:          conf.Server.Address,
		DatabaseURI:         conf.Db.DatabaseURI,
		AccrualSystemAdress: conf.Accural.AccuralSystemAdress,
		JWTSecret:           conf.Auth.JWTSecret,
		TokenExp:            conf.Auth.TokenExp,
	}

	// Регистрируем флаги и перезаписываем данные из конфигурации
	fs.StringVar(&flags.RunAddress, "a", flags.RunAddress, "service run address")
	fs.StringVar(&flags.DatabaseURI, "d", flags.DatabaseURI, "postgres connection uri")
	fs.StringVar(&flags.AccrualSystemAdress, "r", flags.AccrualSystemAdress, "accrual system address")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	if flags.DatabaseURI == "" {
		return nil, errors.New("database URI is required")
	}
	if flags.AccrualSystemAdress == "" {
		return nil, errors.New("accrual system address is required")
	}

	return flags, nil
}

// Загружает флаги и конфигурацию
func LoadFlags() (*Flags, error) {
	conf, err := loadConfig()
	if err != nil {
		return nil, err
	}
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	return parseFlags(conf, fs, os.Args[1:])
}
