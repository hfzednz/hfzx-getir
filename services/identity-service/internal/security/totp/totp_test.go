package totp_test

import (
	"testing"
	"time"

	"github.com/nexora/identity-service/internal/security/totp"
)

func TestGenerateAndValidate(t *testing.T) {
	secret, err := totp.GenerateSecret()
	if err != nil {
		t.Fatal(err)
	}
	fixed := time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)
	opts := totp.Options{Now: func() time.Time { return fixed }}

	code, err := totp.GenerateCode(secret, fixed, opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(code) != 6 {
		t.Fatalf("code=%q", code)
	}
	ok, err := totp.ValidateCode(secret, code, opts)
	if err != nil || !ok {
		t.Fatalf("validate: ok=%v err=%v", ok, err)
	}
	ok, err = totp.ValidateCode(secret, "000000", opts)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("wrong code should fail")
	}
}

func TestOTPAuthURI(t *testing.T) {
	uri, err := totp.OTPAuthURI("JBSWY3DPEHPK3PXP", "user@nexora.local", totp.Options{Issuer: "NEXORA"})
	if err != nil {
		t.Fatal(err)
	}
	if uri == "" {
		t.Fatal("empty uri")
	}
}

// RFC 6238 Appendix B test vector (SHA1, 8 digits, secret "12345678901234567890").
func TestRFC6238Vector(t *testing.T) {
	// base32 of ASCII "12345678901234567890"
	secret := "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	opts := totp.Options{Digits: 8, Period: 30}
	// At Unix 59 → 94287082
	code, err := totp.GenerateCode(secret, time.Unix(59, 0).UTC(), opts)
	if err != nil {
		t.Fatal(err)
	}
	if code != "94287082" {
		t.Fatalf("got %s want 94287082", code)
	}
}
