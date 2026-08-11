package domain_test

import (
	"testing"
	"time"

	"github.com/nexora/liveops-service/internal/domain"
)

func TestEvaluateFlagAndAssign(t *testing.T) {
	f := domain.FeatureFlag{
		Key: "new_checkout", Enabled: true, Percentage: 100,
		Rules: []domain.TargetRule{{Kind: "country", Op: "in", Values: []string{"TR"}}},
	}
	ev := domain.EvaluateFlag(f, domain.EvalContext{SubjectID: "u1", Country: "TR"}, nil)
	if !ev.Enabled {
		t.Fatal(ev)
	}
	ev2 := domain.EvaluateFlag(f, domain.EvalContext{SubjectID: "u1", Country: "DE"}, nil)
	if ev2.Enabled {
		t.Fatal("should miss")
	}
	exp := domain.Experiment{
		Key: "home_ab", Status: "running",
		Variants: []domain.Variant{{Key: "control", Weight: 50}, {Key: "treatment", Weight: 50}},
	}
	v, err := domain.AssignVariant(exp, "user-42")
	if err != nil || v == "" {
		t.Fatal(err)
	}
	v2, _ := domain.AssignVariant(exp, "user-42")
	if v != v2 {
		t.Fatal("sticky broken")
	}
	if !domain.Significant(0.10, 0.13, 0.02) {
		t.Fatal("sig")
	}
	now := time.Now().UTC()
	e := domain.LiveOpsEvent{
		Status: "active", StartsAt: now.Add(-time.Hour), EndsAt: now.Add(time.Hour),
		Scopes: []domain.TargetRule{{Kind: "city", Values: []string{"istanbul"}}},
	}
	if !domain.EventActive(e, now, domain.EvalContext{City: "istanbul"}) {
		t.Fatal("event")
	}
}
