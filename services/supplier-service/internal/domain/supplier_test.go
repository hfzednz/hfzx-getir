package domain

import "testing"

func TestMatchInvoiceHint(t *testing.T) {
	ok, score := MatchInvoiceHint(10000, 10000)
	if !ok || score != 1 {
		t.Fatalf("%v %v", ok, score)
	}
	ok, _ = MatchInvoiceHint(10000, 9000)
	if ok {
		t.Fatal("expected mismatch")
	}
}

func TestComputeOverallScore(t *testing.T) {
	sc := Scorecard{DeliveryScore: 1, QualityScore: 1, PriceScore: 1, ComplianceScore: 1, RiskScore: 0}
	if ComputeOverallScore(sc) != 1 {
		t.Fatal(ComputeOverallScore(sc))
	}
}
