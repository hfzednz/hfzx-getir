package domain

import (
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// ConsentChannel is a marketing / privacy / transactional channel.
type ConsentChannel string

const (
	ConsentChannelEmail         ConsentChannel = "email"
	ConsentChannelSMS           ConsentChannel = "sms"
	ConsentChannelPush          ConsentChannel = "push"
	ConsentChannelWhatsApp      ConsentChannel = "whatsapp"
	ConsentChannelMarketing     ConsentChannel = "marketing"
	ConsentChannelTransactional ConsentChannel = "transactional"
	ConsentChannelPrivacy       ConsentChannel = "privacy"
	ConsentChannelNewsletter    ConsentChannel = "newsletter"
)

func (c ConsentChannel) Valid() bool {
	switch c {
	case ConsentChannelEmail, ConsentChannelSMS, ConsentChannelPush, ConsentChannelWhatsApp,
		ConsentChannelMarketing, ConsentChannelTransactional, ConsentChannelPrivacy, ConsentChannelNewsletter:
		return true
	default:
		return false
	}
}

const maxConsentSourceLen = 64

// Consent is the latest grant/revoke state for a profile channel.
type Consent struct {
	ID         uuid.UUID
	ProfileID  uuid.UUID
	TenantID   uuid.UUID
	Channel    ConsentChannel
	Granted    bool
	Source     string
	RecordedAt time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Validate checks structural invariants.
func (c Consent) Validate() error {
	if c.ID == uuid.Nil {
		return fmt.Errorf("%w: consent id required", ErrInvalidArgument)
	}
	if c.ProfileID == uuid.Nil {
		return fmt.Errorf("%w: profile_id required", ErrInvalidArgument)
	}
	if c.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if !c.Channel.Valid() {
		return fmt.Errorf("%w: invalid consent channel %q", ErrInvalidArgument, c.Channel)
	}
	if utf8.RuneCountInString(c.Source) > maxConsentSourceLen {
		return fmt.Errorf("%w: source too long", ErrInvalidArgument)
	}
	if c.RecordedAt.IsZero() {
		return fmt.Errorf("%w: recorded_at required", ErrInvalidArgument)
	}
	return nil
}
