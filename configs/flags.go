package config

import (
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
	flags := &Flags{}
	fs.StringVar(&flags.RunAddress, "a", ":8080", "service run address")
	fs.StringVar(&flags.DatabaseURI, "d", "", "postgres connection uri")
	fs.StringVar(&flags.AccrualSystemAdress, "r", "", "accrual system address")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if conf.Server.Address != nil {
		flags.RunAddress = *conf.Server.Address
	}
	if conf.Db.DatabaseURI != nil {
		flags.DatabaseURI = *conf.Db.DatabaseURI
	}
	if conf.Accural.AccuralSystemAdress != nil {
		flags.AccrualSystemAdress = *conf.Accural.AccuralSystemAdress
	}

	flags.JWTSecret = conf.Auth.JWTSecret
	flags.TokenExp = conf.Auth.TokenExp

	return flags, nil
}

func LoadFlags() (*Flags, error) {
	conf, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	return parseFlags(conf, fs, os.Args[1:])
}
