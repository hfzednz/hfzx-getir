package domain

import "testing"

func TestReleaseScore(t *testing.T) {
	s := ReleaseScore(true, true, true, 0.1)
	if s < 95 {
		t.Fatal(s)
	}
}

func TestDefaultGraph(t *testing.T) {
	if len(DefaultDependencyGraph()) < 10 {
		t.Fatal("graph")
	}
	if len(GenesisGatesRequired()) < 9 {
		t.Fatal("gates")
	}
}
