package domain_test

import (
	"errors"
	"testing"

	"github.com/nexora/catalog-service/internal/domain"
)

func TestProductStatusTransitions(t *testing.T) {
	tests := []struct {
		from    domain.ProductStatus
		to      domain.ProductStatus
		allowed bool
	}{
		{domain.ProductStatusDraft, domain.ProductStatusPendingReview, true},
		{domain.ProductStatusDraft, domain.ProductStatusPublished, false},
		{domain.ProductStatusPendingReview, domain.ProductStatusApproved, true},
		{domain.ProductStatusPendingReview, domain.ProductStatusDraft, true},
		{domain.ProductStatusApproved, domain.ProductStatusPublished, true},
		{domain.ProductStatusPublished, domain.ProductStatusHidden, true},
		{domain.ProductStatusPublished, domain.ProductStatusDraft, false},
		{domain.ProductStatusHidden, domain.ProductStatusPublished, true},
		{domain.ProductStatusArchived, domain.ProductStatusDraft, true},
		{domain.ProductStatusDeleted, domain.ProductStatusDraft, false},
	}

	for _, tt := range tests {
		name := string(tt.from) + "_to_" + string(tt.to)
		t.Run(name, func(t *testing.T) {
			got := tt.from.CanTransitionTo(tt.to)
			if got != tt.allowed {
				t.Fatalf("CanTransitionTo(%s, %s) = %v, want %v", tt.from, tt.to, got, tt.allowed)
			}
			err := domain.ValidateProductStatusTransition(tt.from, tt.to)
			if tt.allowed && tt.from != tt.to && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tt.allowed && tt.from != tt.to && !errors.Is(err, domain.ErrInvalidTransition) {
				t.Fatalf("expected ErrInvalidTransition, got %v", err)
			}
		})
	}
}

func TestApplyWorkflowAction(t *testing.T) {
	to, err := domain.ApplyWorkflowAction(domain.ProductStatusDraft, domain.ApprovalActionSubmit)
	if err != nil || to != domain.ProductStatusPendingReview {
		t.Fatalf("submit: to=%s err=%v", to, err)
	}
	_, err = domain.ApplyWorkflowAction(domain.ProductStatusDraft, domain.ApprovalActionPublish)
	if !errors.Is(err, domain.ErrInvalidTransition) {
		t.Fatalf("publish from draft should fail: %v", err)
	}
}

func TestValidateBarcodeEAN(t *testing.T) {
	if err := domain.ValidateBarcodeFormat(domain.SKUTypeEAN, "4006381333931"); err != nil {
		t.Fatalf("valid EAN rejected: %v", err)
	}
	if err := domain.ValidateBarcodeFormat(domain.SKUTypeEAN, "4006381333930"); err == nil {
		t.Fatal("invalid check digit should fail")
	}
}
