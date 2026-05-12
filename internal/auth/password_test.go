package auth

import "testing"

func TestHashPassword_VerifyPassword(t *testing.T) {
	h, err := HashPassword("hello")
	if err != nil {
		t.Fatal(err)
	}
	if !VerifyPassword(h, "hello") {
		t.Fatal("verify ok")
	}
	if VerifyPassword(h, "other") {
		t.Fatal("verify fail expected")
	}
}
