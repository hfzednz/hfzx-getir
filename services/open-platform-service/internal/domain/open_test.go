package domain

import "testing"

func TestSignWebhook(t *testing.T) {
	sig := SignWebhook("sec", `{"a":1}`)
	if !VerifyWebhook("sec", `{"a":1}`, sig) {
		t.Fatal(sig)
	}
	if VerifyWebhook("sec", `{"a":2}`, sig) {
		t.Fatal("should fail")
	}
}

func TestHashSecret(t *testing.T) {
	if HashSecret("a") == HashSecret("b") {
		t.Fatal()
	}
}

func TestCatalogDefaults(t *testing.T) {
	if len(DefaultPublicCatalog()) < 10 {
		t.Fatal()
	}
}
