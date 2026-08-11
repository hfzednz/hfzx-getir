package refresh_test

import (
	"testing"

	"github.com/nexora/identity-service/internal/security/refresh"
)

func TestGenerateAndHash(t *testing.T) {
	tok, err := refresh.Generate("")
	if err != nil {
		t.Fatal(err)
	}
	if tok.Raw == "" || tok.Hash == "" || tok.FamilyID == "" {
		t.Fatalf("incomplete token: %+v", tok)
	}
	h, err := refresh.Hash(tok.Raw)
	if err != nil {
		t.Fatal(err)
	}
	if h != tok.Hash {
		t.Fatalf("hash mismatch")
	}
	ok, err := refresh.Matches(tok.Raw, tok.Hash)
	if err != nil || !ok {
		t.Fatalf("matches: ok=%v err=%v", ok, err)
	}

	// rotation within same family
	tok2, err := refresh.Generate(tok.FamilyID)
	if err != nil {
		t.Fatal(err)
	}
	if tok2.FamilyID != tok.FamilyID {
		t.Fatal("family should be preserved")
	}
	if tok2.Raw == tok.Raw {
		t.Fatal("raw tokens must differ")
	}
}

func TestNewFamilyID(t *testing.T) {
	a := refresh.NewFamilyID()
	b := refresh.NewFamilyID()
	if a == "" || a == b {
		t.Fatalf("a=%s b=%s", a, b)
	}
}
