package domain

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// SegmentKind classifies how a segment is defined/evaluated.
type SegmentKind string

const (
	SegmentKindDynamic   SegmentKind = "dynamic"
	SegmentKindBehavior  SegmentKind = "behavior"
	SegmentKindLocation  SegmentKind = "location"
	SegmentKindRevenue   SegmentKind = "revenue"
	SegmentKindRetention SegmentKind = "retention"
	SegmentKindAI        SegmentKind = "ai"
)

func (k SegmentKind) Valid() bool {
	switch k {
	case SegmentKindDynamic, SegmentKindBehavior, SegmentKindLocation,
		SegmentKindRevenue, SegmentKindRetention, SegmentKindAI:
		return true
	default:
		return false
	}
}

const (
	maxSegmentNameLen        = 120
	maxSegmentDescriptionLen = 500
	maxSegmentSourceLen      = 64
)

// Segment is a local segment definition (membership cache / dynamic rules).
type Segment struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Name        string
	Kind        SegmentKind
	Description string
	Rules       map[string]any
	Active      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Validate checks structural invariants.
func (s Segment) Validate() error {
	if s.ID == uuid.Nil {
		return fmt.Errorf("%w: segment id required", ErrInvalidArgument)
	}
	if s.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if strings.TrimSpace(s.Name) == "" {
		return fmt.Errorf("%w: segment name required", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(s.Name) > maxSegmentNameLen {
		return fmt.Errorf("%w: segment name too long", ErrInvalidArgument)
	}
	if !s.Kind.Valid() {
		return fmt.Errorf("%w: invalid segment kind %q", ErrInvalidArgument, s.Kind)
	}
	if utf8.RuneCountInString(s.Description) > maxSegmentDescriptionLen {
		return fmt.Errorf("%w: segment description too long", ErrInvalidArgument)
	}
	return nil
}

// SegmentMember is a profile membership in a segment.
type SegmentMember struct {
	SegmentID uuid.UUID
	ProfileID uuid.UUID
	JoinedAt  time.Time
	ExpiresAt *time.Time
	Source    string
}

// SegmentMembership is an alias used by application ports.
type SegmentMembership = SegmentMember

// Validate checks structural invariants.
func (m SegmentMember) Validate() error {
	if m.SegmentID == uuid.Nil {
		return fmt.Errorf("%w: segment_id required", ErrInvalidArgument)
	}
	if m.ProfileID == uuid.Nil {
		return fmt.Errorf("%w: profile_id required", ErrInvalidArgument)
	}
	if m.JoinedAt.IsZero() {
		return fmt.Errorf("%w: joined_at required", ErrInvalidArgument)
	}
	if m.ExpiresAt != nil && m.ExpiresAt.Before(m.JoinedAt) {
		return fmt.Errorf("%w: expires_at before joined_at", ErrInvariant)
	}
	if utf8.RuneCountInString(m.Source) > maxSegmentSourceLen {
		return fmt.Errorf("%w: source too long", ErrInvalidArgument)
	}
	return nil
}

// IsExpired reports whether membership has expired.
func (m SegmentMember) IsExpired(now time.Time) bool {
	return m.ExpiresAt != nil && !m.ExpiresAt.After(now)
}
