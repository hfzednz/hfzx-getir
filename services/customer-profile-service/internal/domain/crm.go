package domain

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	maxCRMNoteBodyLen      = 8000
	maxTimelineTypeLen     = 64
)

// CRMNote is an internal agent/CSR note on a customer profile.
type CRMNote struct {
	ID        uuid.UUID
	ProfileID uuid.UUID
	TenantID  uuid.UUID
	AuthorID  uuid.UUID
	Body      string
	Pinned    bool
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// Validate checks structural invariants.
func (n CRMNote) Validate() error {
	if n.ID == uuid.Nil {
		return fmt.Errorf("%w: crm note id required", ErrInvalidArgument)
	}
	if n.ProfileID == uuid.Nil {
		return fmt.Errorf("%w: profile_id required", ErrInvalidArgument)
	}
	if n.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if n.AuthorID == uuid.Nil {
		return fmt.Errorf("%w: author_id required", ErrInvalidArgument)
	}
	if strings.TrimSpace(n.Body) == "" {
		return fmt.Errorf("%w: note body required", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(n.Body) > maxCRMNoteBodyLen {
		return fmt.Errorf("%w: note body too long", ErrInvalidArgument)
	}
	return nil
}

// TimelineEvent is an append-only profile timeline entry.
type TimelineEvent struct {
	ID         uuid.UUID
	ProfileID  uuid.UUID
	TenantID   uuid.UUID
	Type       string
	Payload    map[string]any
	ActorID    *uuid.UUID
	OccurredAt time.Time
	CreatedAt  time.Time
}

// Validate checks structural invariants.
func (e TimelineEvent) Validate() error {
	if e.ID == uuid.Nil {
		return fmt.Errorf("%w: timeline event id required", ErrInvalidArgument)
	}
	if e.ProfileID == uuid.Nil {
		return fmt.Errorf("%w: profile_id required", ErrInvalidArgument)
	}
	if e.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if strings.TrimSpace(e.Type) == "" {
		return fmt.Errorf("%w: timeline event type required", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(e.Type) > maxTimelineTypeLen {
		return fmt.Errorf("%w: timeline event type too long", ErrInvalidArgument)
	}
	if e.ActorID != nil && *e.ActorID == uuid.Nil {
		return fmt.Errorf("%w: actor_id must not be nil uuid", ErrInvalidArgument)
	}
	if e.OccurredAt.IsZero() {
		return fmt.Errorf("%w: occurred_at required", ErrInvalidArgument)
	}
	return nil
}
