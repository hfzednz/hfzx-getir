package domain_test

import (
	"testing"

	"github.com/nexora/erp-service/internal/domain"
)

func TestValidateBalance(t *testing.T) {
	j := domain.JournalEntry{Lines: []domain.JournalLine{
		{AccountCode: "1000", DebitMinor: 100},
		{AccountCode: "2000", CreditMinor: 100},
	}}
	if err := j.ValidateBalance(); err != nil {
		t.Fatal(err)
	}
	j.Lines[1].CreditMinor = 50
	if err := j.ValidateBalance(); err != domain.ErrUnbalanced {
		t.Fatalf("want unbalanced got %v", err)
	}
}

func TestThreeWayMatchAndVAT(t *testing.T) {
	if s := domain.ThreeWayMatchScore(1000, 1000, 1000); s != 1 {
		t.Fatalf("perfect match %v", s)
	}
	if tax := domain.VATAmount(10000, 2000); tax != 2000 {
		t.Fatalf("vat %d", tax)
	}
	if d := domain.StraightLineDepreciation(36000, 36); d != 1000 {
		t.Fatalf("dep %d", d)
	}
}
