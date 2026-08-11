package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/nexora/data-platform-service/internal/domain"
)

func TestAssignAndDecide(t *testing.T) {
	vars := []domain.ExperimentVariant{{Name: "a", Weight: 50}, {Name: "b", Weight: 50}}
	v := domain.AssignVariant(uuid.MustParse("00000000-0000-0000-0000-000000000001"), vars)
	if v != "a" && v != "b" {
		t.Fatal(v)
	}
	if domain.DecideWinner(map[string]float64{"a": 1, "b": 2}) != "b" {
		t.Fatal("winner")
	}
}

func TestSchemaRequiredAndThreshold(t *testing.T) {
	req := domain.RequiredFieldsFromSchema(map[string]any{"required": []any{"orderId", "amountMinor"}})
	if len(req) != 2 {
		t.Fatal(req)
	}
	if err := domain.ValidateRequired(map[string]any{"orderId": "x"}, req); err == nil {
		t.Fatal("expected missing amount")
	}
	if !domain.EvalThreshold(10, "gt", 5) {
		t.Fatal("threshold")
	}
}
