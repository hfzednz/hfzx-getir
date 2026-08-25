package domain

import (
	"errors"
	"testing"
)

func TestValidateTransitionLegalAndIllegal(t *testing.T) {
	if err := ValidateTransition(OrderStatusDraft, OrderStatusPendingPayment); err != nil {
		t.Fatalf("draft → pending_payment should be legal: %v", err)
	}
	if err := ValidateTransition(OrderStatusDraft, OrderStatusDraft); err != nil {
		t.Fatalf("same-status should be legal: %v", err)
	}
	err := ValidateTransition(OrderStatusDraft, OrderStatusDelivered)
	if !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("draft → delivered should be illegal, got %v", err)
	}
	err = ValidateTransition(OrderStatusArchived, OrderStatusDraft)
	if !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("archived → draft should be illegal, got %v", err)
	}
}

func TestOrderTransitionTableCoversOnlyValidStatuses(t *testing.T) {
	for from, tos := range orderTransitions {
		if !from.Valid() {
			t.Fatalf("transition table has unknown from status %q", from)
		}
		for _, to := range tos {
			if !to.Valid() {
				t.Fatalf("transition table %s → unknown %q", from, to)
			}
			if !from.CanTransitionTo(to) {
				t.Fatalf("CanTransitionTo(%s,%s) false despite table entry", from, to)
			}
		}
	}
}

func TestUnknownStatusRejected(t *testing.T) {
	err := ValidateTransition(OrderStatus("not-a-status"), OrderStatusDraft)
	if !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("want invalid argument, got %v", err)
	}
}
