package config

import (
	"testing"
)

func Test_loadConfig_FromEnv(t *testing.T) {
	t.Setenv("RUN_ADDRESS", ":1234")
	t.Setenv("DATABASE_URI", "postgres://localhost/x")
	t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "http://accrual")
	t.Setenv("JWT_SECRET", "secret-x")
	t.Setenv("TOKEN_EXP", "48h")

	cfg, err := loadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.Address != ":1234" {
		t.Fatalf("server address %q", cfg.Server.Address)
	}
	if cfg.Db.DatabaseURI != "postgres://localhost/x" {
		t.Fatal("db uri")
	}
	if cfg.Accural.AccuralSystemAdress != "http://accrual" {
		t.Fatal("accrual")
	}
	if cfg.Auth.JWTSecret != "secret-x" || cfg.Auth.TokenExp != "48h" {
		t.Fatalf("auth %+v", cfg.Auth)
	}
}

func Test_loadConfig_EnvDefaults(t *testing.T) {
	t.Setenv("DATABASE_URI", "postgres://localhost/x")
	t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "http://accrual")

	cfg, err := loadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.Address != ":8080" {
		t.Fatalf("server address %q", cfg.Server.Address)
	}
	if cfg.Auth.JWTSecret != "gophermart-dev-secret" || cfg.Auth.TokenExp != "24h" {
		t.Fatalf("auth %+v", cfg.Auth)
	}
}
