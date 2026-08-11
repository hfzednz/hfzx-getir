package domain

import (
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// PrivacyRequestKind is the type of privacy workflow.
type PrivacyRequestKind string

const (
	PrivacyRequestExport   PrivacyRequestKind = "export"
	PrivacyRequestDelete   PrivacyRequestKind = "delete"
	PrivacyRequestDeletion PrivacyRequestKind = PrivacyRequestDelete // alias
)

func (k PrivacyRequestKind) Valid() bool {
	switch k {
	case PrivacyRequestExport, PrivacyRequestDelete:
		return true
	default:
		return false
	}
}

// PrivacyRequestStatus is the processing state of a privacy request.
type PrivacyRequestStatus string

const (
	PrivacyStatusPending    PrivacyRequestStatus = "pending"
	PrivacyStatusProcessing PrivacyRequestStatus = "processing"
	PrivacyStatusCompleted  PrivacyRequestStatus = "completed"
	PrivacyStatusFailed     PrivacyRequestStatus = "failed"
	PrivacyStatusCancelled  PrivacyRequestStatus = "cancelled"
)

func (s PrivacyRequestStatus) Valid() bool {
	switch s {
	case PrivacyStatusPending, PrivacyStatusProcessing, PrivacyStatusCompleted,
		PrivacyStatusFailed, PrivacyStatusCancelled:
		return true
	default:
		return false
	}
}

const (
	maxPayloadRefLen    = 2048
	maxPrivacyReasonLen = 500
	maxErrorMessageLen  = 1000
)

// PrivacyRequest tracks export/delete privacy workflows for a profile.
type PrivacyRequest struct {
	ID          uuid.UUID
	ProfileID   uuid.UUID
	TenantID    uuid.UUID
	Kind        PrivacyRequestKind
	Status      PrivacyRequestStatus
	PayloadRef  string
	RequestedBy *uuid.UUID
	Reason      string
	ErrorMessage string
	CompletedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Validate checks structural invariants.
func (r PrivacyRequest) Validate() error {
	if r.ID == uuid.Nil {
		return fmt.Errorf("%w: privacy request id required", ErrInvalidArgument)
	}
	if r.ProfileID == uuid.Nil {
		return fmt.Errorf("%w: profile_id required", ErrInvalidArgument)
	}
	if r.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if !r.Kind.Valid() {
		return fmt.Errorf("%w: invalid privacy request kind %q", ErrInvalidArgument, r.Kind)
	}
	if !r.Status.Valid() {
		return fmt.Errorf("%w: invalid privacy request status %q", ErrInvalidArgument, r.Status)
	}
	if utf8.RuneCountInString(r.PayloadRef) > maxPayloadRefLen {
		return fmt.Errorf("%w: payload_ref too long", ErrInvalidArgument)
	}
	if r.RequestedBy != nil && *r.RequestedBy == uuid.Nil {
		return fmt.Errorf("%w: requested_by must not be nil uuid", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(r.Reason) > maxPrivacyReasonLen {
		return fmt.Errorf("%w: reason too long", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(r.ErrorMessage) > maxErrorMessageLen {
		return fmt.Errorf("%w: error_message too long", ErrInvalidArgument)
	}
	if r.Status == PrivacyStatusCompleted && r.CompletedAt == nil {
		return fmt.Errorf("%w: completed privacy request requires completed_at", ErrInvariant)
	}
	return nil
}

// MergeJobStatus is the processing state of a profile merge.
type MergeJobStatus string

const (
	MergeStatusPending   MergeJobStatus = "pending"
	MergeStatusRunning   MergeJobStatus = "running"
	MergeStatusCompleted MergeJobStatus = "completed"
	MergeStatusFailed    MergeJobStatus = "failed"
	MergeStatusCancelled MergeJobStatus = "cancelled"
)

func (s MergeJobStatus) Valid() bool {
	switch s {
	case MergeStatusPending, MergeStatusRunning, MergeStatusCompleted, MergeStatusFailed, MergeStatusCancelled:
		return true
	default:
		return false
	}
}

// MergeJob tracks duplicate detection / profile merge workflows.
type MergeJob struct {
	ID               uuid.UUID
	TenantID         uuid.UUID
	SourceProfileID  uuid.UUID
	TargetProfileID  uuid.UUID
	Status           MergeJobStatus
	DetectionScore   *float64
	DetectionReason  map[string]any
	Result           map[string]any
	RequestedBy      *uuid.UUID
	ErrorMessage     string
	StartedAt        *time.Time
	CompletedAt      *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// Validate checks structural invariants.
func (j MergeJob) Validate() error {
	if j.ID == uuid.Nil {
		return fmt.Errorf("%w: merge job id required", ErrInvalidArgument)
	}
	if j.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if j.SourceProfileID == uuid.Nil {
		return fmt.Errorf("%w: source_profile_id required", ErrInvalidArgument)
	}
	if j.TargetProfileID == uuid.Nil {
		return fmt.Errorf("%w: target_profile_id required", ErrInvalidArgument)
	}
	if j.SourceProfileID == j.TargetProfileID {
		return fmt.Errorf("%w: source and target profiles must differ", ErrInvalidArgument)
	}
	if !j.Status.Valid() {
		return fmt.Errorf("%w: invalid merge job status %q", ErrInvalidArgument, j.Status)
	}
	if j.DetectionScore != nil {
		if *j.DetectionScore < 0 || *j.DetectionScore > 1 {
			return fmt.Errorf("%w: detection_score must be in [0,1]", ErrInvalidArgument)
		}
	}
	if j.RequestedBy != nil && *j.RequestedBy == uuid.Nil {
		return fmt.Errorf("%w: requested_by must not be nil uuid", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(j.ErrorMessage) > maxErrorMessageLen {
		return fmt.Errorf("%w: error_message too long", ErrInvalidArgument)
	}
	if j.Status == MergeStatusCompleted && j.CompletedAt == nil {
		return fmt.Errorf("%w: completed merge job requires completed_at", ErrInvariant)
	}
	return nil
}
