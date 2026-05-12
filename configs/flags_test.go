package config

import (
	"flag"
	"testing"
)

func TestParseFlags_OverrideFromEnv(t *testing.T) {
	addr := ":9090"
	dbu := "postgres://env"
	acc := "http://env"
	conf := &Config{
		Server:  ServerConfig{Address: &addr},
		Db:      DbConfig{DatabaseURI: &dbu},
		Accural: AccuralConfig{AccuralSystemAdress: &acc},
		Auth:    AuthConfig{JWTSecret: "j", TokenExp: "12h"},
	}

	fs := flag.NewFlagSet("t", flag.ContinueOnError)
	flags, err := parseFlags(conf, fs, []string{"-a", ":1111", "-d", "from-flag", "-r", "http://flag"})
	if err != nil {
		t.Fatal(err)
	}
	// env (conf) перекрывает значения флагов
	if flags.RunAddress != ":9090" || flags.DatabaseURI != "postgres://env" || flags.AccrualSystemAdress != "http://env" {
		t.Fatalf("%+v", flags)
	}
	if flags.JWTSecret != "j" || flags.TokenExp != "12h" {
		t.Fatalf("%+v", flags)
	}
}

func TestParseFlags_OnlyFlags(t *testing.T) {
	conf := &Config{Auth: AuthConfig{JWTSecret: "x", TokenExp: "1h"}}
	fs := flag.NewFlagSet("t", flag.ContinueOnError)
	flags, err := parseFlags(conf, fs, []string{"-a", ":2222"})
	if err != nil {
		t.Fatal(err)
	}
	if flags.RunAddress != ":2222" {
		t.Fatal(flags.RunAddress)
	}
}
