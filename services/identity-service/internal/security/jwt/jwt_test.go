package jwt_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	jwtpkg "github.com/nexora/identity-service/internal/security/jwt"
)

func TestGenerateIssueParseJWKS(t *testing.T) {
	km := jwtpkg.NewKeyManager()
	if err := km.Generate(2048); err != nil {
		t.Fatal(err)
	}
	if km.KID() == "" {
		t.Fatal("empty kid")
	}

	claims := jwtpkg.AccessClaims{
		Subject:  "user-1",
		Session:  "sess-1",
		Tenant:   "tenant-a",
		Roles:    []string{"customer"},
		AMR:      []string{"pwd", "otp"},
		ACR:      "aal2",
		DeviceID: "dev-9",
		Issuer:   "https://identity.nexora.local",
		Audience: "nexora-api",
		Expires:  time.Now().UTC().Add(15 * time.Minute),
	}

	token, err := km.IssueAccessToken(claims)
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("not a JWT: %s", token)
	}

	parsed, err := km.ParseAndValidate(token, claims.Issuer, claims.Audience)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Subject != claims.Subject || parsed.Session != claims.Session {
		t.Fatalf("claims mismatch: %+v", parsed)
	}
	if parsed.Tenant != claims.Tenant || parsed.DeviceID != claims.DeviceID {
		t.Fatalf("tid/device mismatch: %+v", parsed)
	}
	if parsed.ACR != claims.ACR {
		t.Fatalf("acr=%s", parsed.ACR)
	}
	if len(parsed.Roles) != 1 || parsed.Roles[0] != "customer" {
		t.Fatalf("roles=%v", parsed.Roles)
	}

	jwks, err := km.ExportJWKS()
	if err != nil {
		t.Fatal(err)
	}
	if len(jwks.Keys) != 1 || jwks.Keys[0].Kid != km.KID() {
		t.Fatalf("jwks=%+v", jwks)
	}
	if jwks.Keys[0].Kty != "RSA" || jwks.Keys[0].Alg != "RS256" {
		t.Fatalf("key meta=%+v", jwks.Keys[0])
	}
	raw, err := km.ExportJWKSJSON()
	if err != nil {
		t.Fatal(err)
	}
	var decoded jwtpkg.JWKS
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
}

func TestIssueMissingClaims(t *testing.T) {
	km := jwtpkg.NewKeyManager()
	_ = km.Generate(2048)

	tests := []struct {
		name   string
		claims jwtpkg.AccessClaims
	}{
		{name: "no sub", claims: jwtpkg.AccessClaims{Session: "s", Expires: time.Now().Add(time.Minute)}},
		{name: "no sid", claims: jwtpkg.AccessClaims{Subject: "u", Expires: time.Now().Add(time.Minute)}},
		{name: "no exp", claims: jwtpkg.AccessClaims{Subject: "u", Session: "s"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := km.IssueAccessToken(tt.claims)
			if !errors.Is(err, jwtpkg.ErrMissingClaims) {
				t.Fatalf("got %v", err)
			}
		})
	}
}

func TestParseRejectsTampered(t *testing.T) {
	km := jwtpkg.NewKeyManager()
	_ = km.Generate(2048)
	token, err := km.IssueAccessToken(jwtpkg.AccessClaims{
		Subject:  "u",
		Session:  "s",
		Issuer:   "iss",
		Audience: "aud",
		Expires:  time.Now().UTC().Add(time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}
	tampered := token + "x"
	_, err = km.ParseAndValidate(tampered, "iss", "aud")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseWrongAudience(t *testing.T) {
	km := jwtpkg.NewKeyManager()
	_ = km.Generate(2048)
	token, err := km.IssueAccessToken(jwtpkg.AccessClaims{
		Subject:  "u",
		Session:  "s",
		Issuer:   "iss",
		Audience: "aud",
		Expires:  time.Now().UTC().Add(time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = km.ParseAndValidate(token, "iss", "other-aud")
	if err == nil {
		t.Fatal("expected audience error")
	}
}

func TestLoadPEMRoundTrip(t *testing.T) {
	km := jwtpkg.NewKeyManager()
	if err := km.Generate(2048); err != nil {
		t.Fatal(err)
	}
	pemBytes, err := km.PrivateKeyPEM()
	if err != nil {
		t.Fatal(err)
	}
	kid := km.KID()

	km2 := jwtpkg.NewKeyManager()
	if err := km2.LoadPEMBytes(pemBytes, kid); err != nil {
		t.Fatal(err)
	}
	token, err := km2.IssueAccessToken(jwtpkg.AccessClaims{
		Subject: "u", Session: "s", Issuer: "iss", Audience: "aud",
		Expires: time.Now().UTC().Add(time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := km.ParseAndValidate(token, "iss", "aud"); err != nil {
		t.Fatal(err)
	}
}

func TestNoKeyErrors(t *testing.T) {
	km := jwtpkg.NewKeyManager()
	_, err := km.IssueAccessToken(jwtpkg.AccessClaims{
		Subject: "u", Session: "s", Expires: time.Now().Add(time.Minute),
	})
	if !errors.Is(err, jwtpkg.ErrNoKey) {
		t.Fatalf("got %v", err)
	}
	_, err = km.ExportJWKS()
	if !errors.Is(err, jwtpkg.ErrNoKey) {
		t.Fatalf("got %v", err)
	}
}
