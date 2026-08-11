package domain

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// TagKind is a built-in or custom customer tag category.
type TagKind string

const (
	TagKindVIP        TagKind = "vip"
	TagKindPremium    TagKind = "premium"
	TagKindNew        TagKind = "new"
	TagKindReturning  TagKind = "returning"
	TagKindHighValue  TagKind = "high_value"
	TagKindInactive   TagKind = "inactive"
	TagKindRisk       TagKind = "risk"
	TagKindFraudWatch TagKind = "fraud_watch"
	TagKindCustom     TagKind = "custom"
)

func (k TagKind) Valid() bool {
	switch k {
	case TagKindVIP, TagKindPremium, TagKindNew, TagKindReturning,
		TagKindHighValue, TagKindInactive, TagKindRisk, TagKindFraudWatch, TagKindCustom:
		return true
	default:
		return false
	}
}

const (
	maxTagNameLen        = 64
	maxTagDescriptionLen = 256
	maxTagColorLen       = 32
	maxTagNoteLen        = 500
)

// Tag is a tenant-scoped tag definition.
type Tag struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Kind        TagKind
	Name        string
	Description string
	Color       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Validate checks structural invariants.
func (t Tag) Validate() error {
	if t.ID == uuid.Nil {
		return fmt.Errorf("%w: tag id required", ErrInvalidArgument)
	}
	if t.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if !t.Kind.Valid() {
		return fmt.Errorf("%w: invalid tag kind %q", ErrInvalidArgument, t.Kind)
	}
	if strings.TrimSpace(t.Name) == "" {
		return fmt.Errorf("%w: tag name required", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(t.Name) > maxTagNameLen {
		return fmt.Errorf("%w: tag name too long", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(t.Description) > maxTagDescriptionLen {
		return fmt.Errorf("%w: tag description too long", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(t.Color) > maxTagColorLen {
		return fmt.Errorf("%w: tag color too long", ErrInvalidArgument)
	}
	return nil
}

// ProfileTag is an assignment of a tag to a profile.
type ProfileTag struct {
	ProfileID  uuid.UUID
	TagID      uuid.UUID
	AssignedBy *uuid.UUID
	AssignedAt time.Time
	Note       string
}

// Validate checks structural invariants.
func (pt ProfileTag) Validate() error {
	if pt.ProfileID == uuid.Nil {
		return fmt.Errorf("%w: profile_id required", ErrInvalidArgument)
	}
	if pt.TagID == uuid.Nil {
		return fmt.Errorf("%w: tag_id required", ErrInvalidArgument)
	}
	if pt.AssignedBy != nil && *pt.AssignedBy == uuid.Nil {
		return fmt.Errorf("%w: assigned_by must not be nil uuid", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(pt.Note) > maxTagNoteLen {
		return fmt.Errorf("%w: note too long", ErrInvalidArgument)
	}
	return nil
}
