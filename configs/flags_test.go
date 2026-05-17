package config

import (
	"flag"
	"testing"
)

func TestParseFlags_FlagsOverrideEnv(t *testing.T) {
	conf := &Config{
		Server:  ServerConfig{Address: ":9090"},
		Db:      DbConfig{DatabaseURI: "postgres://env"},
		Accural: AccuralConfig{AccuralSystemAdress: "http://env"},
		Auth:    AuthConfig{JWTSecret: "j", TokenExp: "12h"},
	}

	fs := flag.NewFlagSet("t", flag.ContinueOnError)
	flags, err := parseFlags(conf, fs, []string{"-a", ":1111", "-d", "from-flag", "-r", "http://flag"})
	if err != nil {
		t.Fatal(err)
	}
	if flags.RunAddress != ":1111" || flags.DatabaseURI != "from-flag" || flags.AccrualSystemAdress != "http://flag" {
		t.Fatalf("%+v", flags)
	}
	if flags.JWTSecret != "j" || flags.TokenExp != "12h" {
		t.Fatalf("%+v", flags)
	}
}

func TestParseFlags_OnlyFlags(t *testing.T) {
	conf := &Config{Auth: AuthConfig{JWTSecret: "x", TokenExp: "1h"}}
	fs := flag.NewFlagSet("t", flag.ContinueOnError)
	flags, err := parseFlags(conf, fs, []string{
		"-a", ":1234",
		"-d", "postgres://db",
		"-r", "http://accrual",
	})
	if err != nil {
		t.Fatal(err)
	}
	if flags.RunAddress != ":1234" || flags.DatabaseURI != "postgres://db" {
		t.Fatal(flags)
	}
}

func TestParseFlags_ConfigDefaults(t *testing.T) {
	conf := &Config{
		Server:  ServerConfig{Address: ":8080"},
		Db:      DbConfig{DatabaseURI: "postgres://env"},
		Accural: AccuralConfig{AccuralSystemAdress: "http://env"},
		Auth:    AuthConfig{JWTSecret: "j", TokenExp: "12h"},
	}

	fs := flag.NewFlagSet("t", flag.ContinueOnError)
	flags, err := parseFlags(conf, fs, nil)
	if err != nil {
		t.Fatal(err)
	}
	if flags.RunAddress != ":8080" || flags.DatabaseURI != "postgres://env" || flags.AccrualSystemAdress != "http://env" {
		t.Fatalf("%+v", flags)
	}
}

func TestParseFlags_MissingRequired(t *testing.T) {
	conf := &Config{Auth: AuthConfig{JWTSecret: "x", TokenExp: "1h"}}
	fs := flag.NewFlagSet("t", flag.ContinueOnError)
	_, err := parseFlags(conf, fs, []string{"-a", ":2222"})
	if err == nil {
		t.Fatal("expected error")
	}
}
