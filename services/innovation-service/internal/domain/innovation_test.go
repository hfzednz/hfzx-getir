package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestValidateModuleAndEnable(t *testing.T) {
	m := InnovationModule{TenantID: mustUUID(), Key: "x", Name: "X", Domain: DomainEdge, TRL: 2}
	if err := ValidateModule(m); err != nil {
		t.Fatal(err)
	}
	if err := CanEnable(m); err != ErrNotReady {
		t.Fatalf("want not ready, got %v", err)
	}
	m.TRL = 7
	m.SandboxOnly = false
	if err := CanEnable(m); err != nil {
		t.Fatal(err)
	}
}

func TestDefaultCatalog(t *testing.T) {
	c := DefaultInnovationCatalog()
	if len(c) < 15 {
		t.Fatal(len(c))
	}
}

func TestEstimateSimAccuracy(t *testing.T) {
	a := EstimateSimAccuracy(SimDemand, map[string]any{"iterations": float64(200)})
	if a < 0.78 {
		t.Fatal(a)
	}
}

func mustUUID() uuid.UUID {
	return uuid.MustParse("11111111-1111-4111-8111-111111111111")
}
