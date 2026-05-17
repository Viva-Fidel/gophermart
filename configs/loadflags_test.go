package config

import (
	"os"
	"testing"
)

func TestLoadFlags_MergesEnvAndArgs(t *testing.T) {
	t.Setenv("JWT_SECRET", "sec")
	t.Setenv("TOKEN_EXP", "2h")
	t.Setenv("RUN_ADDRESS", ":1234")
	t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "http://accrual-env")

	old := os.Args
	t.Cleanup(func() { os.Args = old })
	os.Args = []string{"gophermart", "-d", "postgres://flag"}

	flags, err := LoadFlags()
	if err != nil {
		t.Fatal(err)
	}
	if flags.RunAddress != ":1234" {
		t.Fatal(flags.RunAddress)
	}
	if flags.DatabaseURI != "postgres://flag" {
		t.Fatal(flags.DatabaseURI)
	}
	if flags.JWTSecret != "sec" || flags.TokenExp != "2h" {
		t.Fatal(flags)
	}
}
