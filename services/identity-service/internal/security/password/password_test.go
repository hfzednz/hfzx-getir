package password_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/nexora/identity-service/internal/security/password"
)

func TestHashAndVerify(t *testing.T) {
	h := password.NewDefaultHasher()

	tests := []struct {
		name     string
		password string
		verify   string
		wantOK   bool
		wantErr  bool
	}{
		{name: "match", password: "CorrectHorseBattery!", verify: "CorrectHorseBattery!", wantOK: true},
		{name: "mismatch", password: "CorrectHorseBattery!", verify: "wrong-password", wantOK: false},
		{name: "empty", password: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := h.Hash(tt.password)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("Hash: %v", err)
			}
			if !strings.HasPrefix(encoded, "$argon2id$") {
				t.Fatalf("unexpected encoding: %s", encoded)
			}
			ok, err := h.Verify(tt.verify, encoded)
			if err != nil {
				t.Fatalf("Verify: %v", err)
			}
			if ok != tt.wantOK {
				t.Fatalf("got ok=%v want %v", ok, tt.wantOK)
			}
		})
	}
}

func TestValidateComplexity(t *testing.T) {
	policy := password.ComplexityPolicy{
		MinLength:     8,
		RequireUpper:  true,
		RequireLower:  true,
		RequireDigit:  true,
		RequireSymbol: true,
		MaxLength:     64,
	}

	tests := []struct {
		name    string
		pw      string
		wantErr bool
	}{
		{name: "valid", pw: "Abcdef1!", wantErr: false},
		{name: "too short", pw: "Ab1!", wantErr: true},
		{name: "no upper", pw: "abcdef1!", wantErr: true},
		{name: "no lower", pw: "ABCDEF1!", wantErr: true},
		{name: "no digit", pw: "Abcdefg!", wantErr: true},
		{name: "no symbol", pw: "Abcdefg1", wantErr: true},
		{name: "too long", pw: strings.Repeat("Aa1!", 20), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := password.ValidateComplexity(tt.pw, policy)
			if tt.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantErr && err != nil && !errors.Is(err, password.ErrComplexityFailed) {
				t.Fatalf("want ErrComplexityFailed, got %v", err)
			}
		})
	}
}

func TestCheckHistory(t *testing.T) {
	h := password.NewDefaultHasher()
	old1, err := h.Hash("OldPassword1!")
	if err != nil {
		t.Fatal(err)
	}
	old2, err := h.Hash("OldPassword2!")
	if err != nil {
		t.Fatal(err)
	}
	history := []string{old1, old2}

	tests := []struct {
		name    string
		pw      string
		wantErr error
	}{
		{name: "reused first", pw: "OldPassword1!", wantErr: password.ErrPasswordReused},
		{name: "reused second", pw: "OldPassword2!", wantErr: password.ErrPasswordReused},
		{name: "fresh", pw: "BrandNewPass3!", wantErr: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := password.CheckHistory(tt.pw, history, h)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("got %v want %v", err, tt.wantErr)
			}
		})
	}
}

func TestBreachCheckMock(t *testing.T) {
	checker := password.NewMockBreachChecker("password", "123456")

	tests := []struct {
		name    string
		pw      string
		wantErr error
	}{
		{name: "breached", pw: "password", wantErr: password.ErrPasswordBreached},
		{name: "clean", pw: "UniqueEnough!9", wantErr: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := password.EnsureNotBreached(tt.pw, checker)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("got %v want %v", err, tt.wantErr)
			}
		})
	}
}

func TestVerifyInvalidHash(t *testing.T) {
	h := password.NewDefaultHasher()
	_, err := h.Verify("anything", "not-a-valid-hash")
	if !errors.Is(err, password.ErrInvalidHash) {
		t.Fatalf("got %v want ErrInvalidHash", err)
	}
}
