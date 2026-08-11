package policy_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/nexora/identity-service/internal/domain"
	"github.com/nexora/identity-service/internal/domain/policy"
)

func TestExpandRoles_Inheritance(t *testing.T) {
	parentID := uuid.New()
	childID := uuid.New()
	permParent := domain.Permission{Resource: "admin", Action: "read"}
	permChild := domain.Permission{Resource: "orders", Action: "write"}

	graph := policy.RoleGraph{
		parentID: {
			Role:        domain.Role{ID: parentID, Name: "admin"},
			Permissions: []domain.Permission{permParent},
		},
		childID: {
			Role:        domain.Role{ID: childID, Name: "city_ops"},
			ParentIDs:   []uuid.UUID{parentID},
			Permissions: []domain.Permission{permChild},
		},
	}

	got, err := policy.ExpandRoles(graph, []uuid.UUID{childID})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 perms, got %d", len(got))
	}
}

func TestExpandRoles_Cycle(t *testing.T) {
	a := uuid.New()
	b := uuid.New()
	graph := policy.RoleGraph{
		a: {Role: domain.Role{ID: a, Name: "a"}, ParentIDs: []uuid.UUID{b}},
		b: {Role: domain.Role{ID: b, Name: "b"}, ParentIDs: []uuid.UUID{a}},
	}
	if _, err := policy.ExpandRoles(graph, []uuid.UUID{a}); err == nil {
		t.Fatal("expected cycle error")
	}
}

func TestABAC_TenantMismatch(t *testing.T) {
	tid := uuid.New()
	other := uuid.New()
	attrs := policy.Attributes{TenantID: &tid, RiskScore: 10}
	req := policy.Requirement{RequiredTenant: &other, MaxRisk: 100}
	dec := policy.Evaluate(attrs, req)
	if dec.Allow {
		t.Fatal("expected deny on tenant mismatch")
	}
}

func TestABAC_Allow(t *testing.T) {
	tid := uuid.New()
	attrs := policy.Attributes{TenantID: &tid, RiskScore: 10, MFALevel: 1}
	req := policy.Requirement{RequiredTenant: &tid, MaxRisk: 50, MinMFALevel: 1}
	dec := policy.Evaluate(attrs, req)
	if !dec.Allow {
		t.Fatalf("expected allow, reasons=%v", dec.Reasons)
	}
}
