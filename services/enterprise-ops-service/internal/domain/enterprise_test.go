package domain

import "testing"

func TestRiskScore(t *testing.T) {
	s, err := RiskScore(3, 4)
	if err != nil || s != 12 {
		t.Fatalf("%d %v", s, err)
	}
	if _, err := RiskScore(0, 2); err != ErrInvalidArgument {
		t.Fatal(err)
	}
}

func TestProjectHealth(t *testing.T) {
	if ProjectHealth(0.5, 0) != "green" {
		t.Fatal("green")
	}
	if ProjectHealth(0.95, 1) != "amber" {
		t.Fatal("amber")
	}
	if ProjectHealth(1.2, 0) != "red" {
		t.Fatal("red")
	}
}
