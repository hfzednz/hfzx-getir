package domain

import "testing"

func TestEvaluateUnitCoverage(t *testing.T) {
	ok, _ := EvaluateUnitCoverage(80, 70)
	if !ok {
		t.Fatal()
	}
	ok, _ = EvaluateUnitCoverage(60, 70)
	if ok {
		t.Fatal()
	}
}

func TestEvaluatePerf(t *testing.T) {
	ok, _ := EvaluatePerf(200, 500, 0.001, 0.01)
	if !ok {
		t.Fatal()
	}
	ok, _ = EvaluatePerf(800, 500, 0, 0.01)
	if ok {
		t.Fatal()
	}
}

func TestEvaluateSecurity(t *testing.T) {
	ok, _ := EvaluateSecurity(1, 0, false)
	if ok {
		t.Fatal()
	}
	ok, _ = EvaluateSecurity(0, 0, false)
	if !ok {
		t.Fatal()
	}
}

func TestAllRequiredGatesPassed(t *testing.T) {
	pols := []GatePolicy{{Key: "unit", Required: true}, {Key: "sec", Required: true}}
	evals := []GateEvaluation{{PolicyKey: "unit", Passed: true}, {PolicyKey: "sec", Passed: true}}
	if !AllRequiredGatesPassed(evals, pols) {
		t.Fatal()
	}
	evals[1].Passed = false
	if AllRequiredGatesPassed(evals, pols) {
		t.Fatal()
	}
}
