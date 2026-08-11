package domain

import "testing"

func TestBenchmarkPasses(t *testing.T) {
	if !BenchmarkPasses(BenchOrders, 5000, 5000) {
		t.Fatal("orders")
	}
	if !BenchmarkPasses(BenchAPILatency, 120, 150) {
		t.Fatal("latency")
	}
	if BenchmarkPasses(BenchAPILatency, 200, 150) {
		t.Fatal("latency fail")
	}
}

func TestDefaultTargets(t *testing.T) {
	if len(DefaultTargets()) < 8 {
		t.Fatal("targets")
	}
	if len(DefaultCapacityCatalog()) < 5 {
		t.Fatal("capacity")
	}
}
