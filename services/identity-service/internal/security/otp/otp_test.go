package otp_test

import (
	"errors"
	"testing"
	"time"
	"unicode"

	"github.com/nexora/identity-service/internal/security/otp"
)

func TestGenerateNumeric(t *testing.T) {
	tests := []struct {
		name    string
		n       int
		wantErr bool
	}{
		{name: "six", n: 6},
		{name: "four", n: 4},
		{name: "zero", n: 0, wantErr: true},
		{name: "too large", n: 99, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := otp.GenerateNumeric(tt.n)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(code) != tt.n {
				t.Fatalf("len=%d want %d", len(code), tt.n)
			}
			for _, r := range code {
				if !unicode.IsDigit(r) {
					t.Fatalf("non-digit in %q", code)
				}
			}
		})
	}
}

func TestHashAndVerifyOTP(t *testing.T) {
	pepper := "server-pepper"

	tests := []struct {
		name   string
		otp    string
		verify string
		want   bool
	}{
		{name: "match", otp: "123456", verify: "123456", want: true},
		{name: "mismatch", otp: "123456", verify: "654321", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := otp.HashOTP(tt.otp, pepper)
			if err != nil {
				t.Fatal(err)
			}
			ok, err := otp.VerifyOTP(tt.verify, pepper, hash)
			if err != nil {
				t.Fatal(err)
			}
			if ok != tt.want {
				t.Fatalf("got %v want %v", ok, tt.want)
			}
		})
	}
}

func TestVerifyOTPEmpty(t *testing.T) {
	_, err := otp.VerifyOTP("", "p", "hash")
	if !errors.Is(err, otp.ErrEmptyOTP) {
		t.Fatalf("got %v", err)
	}
}

func TestTTLHelpers(t *testing.T) {
	exp := otp.ExpiresAt(100 * time.Millisecond)
	if otp.IsExpired(exp) {
		t.Fatal("should not be expired yet")
	}
	if rem := otp.Remaining(exp); rem <= 0 {
		t.Fatalf("remaining=%v", rem)
	}
	time.Sleep(150 * time.Millisecond)
	if !otp.IsExpired(exp) {
		t.Fatal("expected expired")
	}
	if err := otp.EnsureValid(exp); !errors.Is(err, otp.ErrExpired) {
		t.Fatalf("got %v", err)
	}
	if rem := otp.Remaining(exp); rem != 0 {
		t.Fatalf("remaining after expiry=%v", rem)
	}
}

func TestChallengeFlow(t *testing.T) {
	plain, ch, err := otp.NewChallenge(6, "pepper", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if len(plain) != 6 {
		t.Fatalf("plain len=%d", len(plain))
	}

	ok, err := otp.VerifyChallenge(plain, "pepper", ch)
	if err != nil || !ok {
		t.Fatalf("verify: ok=%v err=%v", ok, err)
	}

	ok, err = otp.VerifyChallenge("000000", "pepper", ch)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("wrong otp should fail")
	}

	expired := ch
	expired.ExpiresAt = time.Now().UTC().Add(-time.Second)
	_, err = otp.VerifyChallenge(plain, "pepper", expired)
	if !errors.Is(err, otp.ErrExpired) {
		t.Fatalf("got %v", err)
	}
}
