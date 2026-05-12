package config

import (
	"testing"
)

func TestLoadConfig_FromEnv(t *testing.T) {
	t.Setenv("RUN_ADDRESS", ":1234")
	t.Setenv("DATABASE_URI", "postgres://localhost/x")
	t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "http://accrual")
	t.Setenv("JWT_SECRET", "secret-x")
	t.Setenv("TOKEN_EXP", "48h")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.Address == nil || *cfg.Server.Address != ":1234" {
		t.Fatalf("server address %+v", cfg.Server.Address)
	}
	if cfg.Db.DatabaseURI == nil || *cfg.Db.DatabaseURI != "postgres://localhost/x" {
		t.Fatal("db uri")
	}
	if cfg.Accural.AccuralSystemAdress == nil || *cfg.Accural.AccuralSystemAdress != "http://accrual" {
		t.Fatal("accrual")
	}
	if cfg.Auth.JWTSecret != "secret-x" || cfg.Auth.TokenExp != "48h" {
		t.Fatalf("auth %+v", cfg.Auth)
	}
}
