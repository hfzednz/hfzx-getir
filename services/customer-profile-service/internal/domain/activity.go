package domain

import (
	"fmt"
	"net"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	maxActivityActionLen       = 64
	maxActivityResourceTypeLen = 64
	maxActivityUserAgentLen    = 512
)

// ActivityEntry is a profile-side activity log record.
type ActivityEntry struct {
	ID           uuid.UUID
	ProfileID    uuid.UUID
	TenantID     uuid.UUID
	ActorID      *uuid.UUID
	Action       string
	ResourceType string
	ResourceID   *uuid.UUID
	Payload      map[string]any
	IP           net.IP
	UserAgent    string
	OccurredAt   time.Time
	CreatedAt    time.Time
}

// Validate checks structural invariants.
func (a ActivityEntry) Validate() error {
	if a.ID == uuid.Nil {
		return fmt.Errorf("%w: activity id required", ErrInvalidArgument)
	}
	if a.ProfileID == uuid.Nil {
		return fmt.Errorf("%w: profile_id required", ErrInvalidArgument)
	}
	if a.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if strings.TrimSpace(a.Action) == "" {
		return fmt.Errorf("%w: action required", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(a.Action) > maxActivityActionLen {
		return fmt.Errorf("%w: action too long", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(a.ResourceType) > maxActivityResourceTypeLen {
		return fmt.Errorf("%w: resource_type too long", ErrInvalidArgument)
	}
	if a.ActorID != nil && *a.ActorID == uuid.Nil {
		return fmt.Errorf("%w: actor_id must not be nil uuid", ErrInvalidArgument)
	}
	if a.ResourceID != nil && *a.ResourceID == uuid.Nil {
		return fmt.Errorf("%w: resource_id must not be nil uuid", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(a.UserAgent) > maxActivityUserAgentLen {
		return fmt.Errorf("%w: user_agent too long", ErrInvalidArgument)
	}
	if a.OccurredAt.IsZero() {
		return fmt.Errorf("%w: occurred_at required", ErrInvalidArgument)
	}
	return nil
}

// ActivityOrdersSummary is a compact orders/activity signal used to recompute AI scores.
// Sourced from projections / analytics — not stored as a first-class table here.
type ActivityOrdersSummary struct {
	OrderCount30d  int
	OrderCount90d  int
	AvgOrderValue  float64
	DaysSinceLast  int
	CancelRate     float64
	PreferredCats  []string
}
