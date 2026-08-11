package domain

import (
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// ApprovalActionType is a workflow audit action.
type ApprovalActionType string

const (
	ApprovalActionSubmit         ApprovalActionType = "submit"
	ApprovalActionApprove        ApprovalActionType = "approve"
	ApprovalActionReject         ApprovalActionType = "reject"
	ApprovalActionRequestChanges ApprovalActionType = "request_changes"
	ApprovalActionPublish        ApprovalActionType = "publish"
	ApprovalActionUnpublish      ApprovalActionType = "unpublish"
	ApprovalActionSchedule       ApprovalActionType = "schedule"
	ApprovalActionRollback       ApprovalActionType = "rollback"
	ApprovalActionArchive        ApprovalActionType = "archive"
)

func (a ApprovalActionType) Valid() bool {
	switch a {
	case ApprovalActionSubmit, ApprovalActionApprove, ApprovalActionReject,
		ApprovalActionRequestChanges, ApprovalActionPublish, ApprovalActionUnpublish,
		ApprovalActionSchedule, ApprovalActionRollback, ApprovalActionArchive:
		return true
	default:
		return false
	}
}

// TargetStatus returns the intended product status after a successful action, if known.
func (a ApprovalActionType) TargetStatus() (ProductStatus, bool) {
	switch a {
	case ApprovalActionSubmit:
		return ProductStatusPendingReview, true
	case ApprovalActionApprove:
		return ProductStatusApproved, true
	case ApprovalActionReject, ApprovalActionRequestChanges:
		return ProductStatusDraft, true
	case ApprovalActionPublish:
		return ProductStatusPublished, true
	case ApprovalActionUnpublish:
		return ProductStatusHidden, true
	case ApprovalActionSchedule:
		return ProductStatusScheduled, true
	case ApprovalActionArchive:
		return ProductStatusArchived, true
	default:
		return "", false
	}
}

const maxCommentLen = 2000

// ApprovalAction is an immutable workflow audit record.
type ApprovalAction struct {
	ID         uuid.UUID
	ProductID  uuid.UUID
	VersionID  *uuid.UUID
	TenantID   uuid.UUID
	Action     ApprovalActionType
	FromStatus *ProductStatus
	ToStatus   *ProductStatus
	ActorID    uuid.UUID
	ActorRole  string
	Comment    string
	Metadata   map[string]any
	CreatedAt  time.Time
}

// Validate checks structural invariants.
func (a ApprovalAction) Validate() error {
	if a.ID == uuid.Nil {
		return fmt.Errorf("%w: approval_action id required", ErrInvalidArgument)
	}
	if a.ProductID == uuid.Nil {
		return fmt.Errorf("%w: product_id required", ErrInvalidArgument)
	}
	if a.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if !a.Action.Valid() {
		return fmt.Errorf("%w: invalid approval action %q", ErrInvalidArgument, a.Action)
	}
	if a.ActorID == uuid.Nil {
		return fmt.Errorf("%w: actor_id required", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(a.Comment) > maxCommentLen {
		return fmt.Errorf("%w: comment too long", ErrInvalidArgument)
	}
	if a.FromStatus != nil && !a.FromStatus.Valid() {
		return fmt.Errorf("%w: invalid from_status %q", ErrInvalidArgument, *a.FromStatus)
	}
	if a.ToStatus != nil && !a.ToStatus.Valid() {
		return fmt.Errorf("%w: invalid to_status %q", ErrInvalidArgument, *a.ToStatus)
	}
	if a.FromStatus != nil && a.ToStatus != nil {
		if err := ValidateProductStatusTransition(*a.FromStatus, *a.ToStatus); err != nil {
			return err
		}
	}
	return nil
}

// ApplyWorkflowAction derives the next product status for a workflow action.
func ApplyWorkflowAction(current ProductStatus, action ApprovalActionType) (ProductStatus, error) {
	target, ok := action.TargetStatus()
	if !ok {
		// rollback keeps caller-specified status; validate separately
		if action == ApprovalActionRollback {
			return current, nil
		}
		return "", fmt.Errorf("%w: unsupported action %q", ErrInvalidArgument, action)
	}
	if err := ValidateProductStatusTransition(current, target); err != nil {
		return "", err
	}
	return target, nil
}
