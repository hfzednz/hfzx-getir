package domain

import (
	"time"

	"github.com/google/uuid"
)

// Preference holds per-principal channel opt-in/out and quiet hours.
type Preference struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	PrincipalID uuid.UUID
	// ChannelOptOut maps channel → opted out (true = do not send marketing/non-override).
	ChannelOptOut map[Channel]bool
	QuietStart    int // hour 0-23 inclusive start (local conceptually; UTC in memory mode)
	QuietEnd      int // hour 0-23 exclusive end; equal means disabled
	UpdatedAt     time.Time
}

// DefaultPreference returns allow-all prefs with quiet hours disabled.
func DefaultPreference(tenantID, principalID uuid.UUID, now time.Time) Preference {
	return Preference{
		TenantID:      tenantID,
		PrincipalID:   principalID,
		ChannelOptOut: map[Channel]bool{},
		QuietStart:    0,
		QuietEnd:      0,
		UpdatedAt:     now,
	}
}

// Allow decides whether a send is permitted for channel/priority at now.
// Transactional/otp/system override marketing opt-out. Quiet hours suppress marketing only.
func (p Preference) Allow(channel Channel, priority Priority, now time.Time) bool {
	if p.ChannelOptOut == nil {
		p.ChannelOptOut = map[Channel]bool{}
	}
	optedOut := p.ChannelOptOut[channel]
	if optedOut && !priority.OverridesMarketingOptOut() {
		return false
	}
	if priority.RespectsQuietHours() && inQuietHours(now, p.QuietStart, p.QuietEnd) {
		return false
	}
	return true
}

// inQuietHours returns true when quiet hours are configured and now falls inside the window.
// QuietEnd == QuietStart means disabled. Window may wrap midnight (e.g. 22→7).
func inQuietHours(now time.Time, start, end int) bool {
	if start == end {
		return false
	}
	if start < 0 || start > 23 || end < 0 || end > 23 {
		return false
	}
	h := now.UTC().Hour()
	if start < end {
		return h >= start && h < end
	}
	// wraps midnight
	return h >= start || h < end
}

// Consent records GDPR/KVKK style consent.
type Consent struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	PrincipalID uuid.UUID
	Purpose     string
	Granted     bool
	Source      string
	RecordedAt  time.Time
}
