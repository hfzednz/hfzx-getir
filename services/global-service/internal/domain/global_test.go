package domain

import "testing"

func TestConvertMinor(t *testing.T) {
	// 100.00 TRY -> EUR at 0.03 => 3.00 EUR (minor 300 if both 2)
	got := ConvertMinor(10000, 0.03, 2, 2)
	if got != 300 {
		t.Fatalf("got %d", got)
	}
}

func TestTaxAmountMinor(t *testing.T) {
	if TaxAmountMinor(10000, 1800) != 1800 {
		t.Fatal(TaxAmountMinor(10000, 1800))
	}
}

func TestResolveTranslationFallback(t *testing.T) {
	v, ok := ResolveTranslation(map[string]string{}, map[string]string{"a": "A"}, "a")
	if !ok || v != "A" {
		t.Fatal(v, ok)
	}
}
