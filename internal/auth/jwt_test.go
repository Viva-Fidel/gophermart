package auth

import (
	"testing"
	"time"
)

func TestJWT_RoundTrip(t *testing.T) {
	secret := "jwt-test"
	tok, err := NewToken(99, secret, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	cl, err := ParseToken(tok, secret)
	if err != nil {
		t.Fatal(err)
	}
	if cl.UserID != 99 {
		t.Fatal(cl.UserID)
	}
}

func TestParseToken_BadSecret(t *testing.T) {
	tok, err := NewToken(1, "123", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	_, err = ParseToken(tok, "321")
	if err == nil {
		t.Fatal("want error")
	}
}
