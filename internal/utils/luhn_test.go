package utils

import "testing"

func TestValidLuhn(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"", false},
		{"12a", false},
		{"79927398713", true},
		{"79927398714", false},
		{"0", true},
		{"18", true},
	}
	for _, tt := range tests {
		if got := ValidLuhn(tt.in); got != tt.want {
			t.Fatalf("ValidLuhn(%q)=%v want %v", tt.in, got, tt.want)
		}
	}
}
