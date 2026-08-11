package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// SessionStatus is the checkout session lifecycle status.
type SessionStatus string

const (
	StatusStarted     SessionStatus = "started"
	StatusValidating  SessionStatus = "validating"
	StatusReady       SessionStatus = "ready"
	StatusBlocked     SessionStatus = "blocked"
	StatusCompleting  SessionStatus = "completing"
	StatusCompleted   SessionStatus = "completed"
	StatusFailed      SessionStatus = "failed"
	StatusAbandoned   SessionStatus = "abandoned"
)

// Valid reports whether the status is recognized.
func (s SessionStatus) Valid() bool {
	switch s {
	case StatusStarted, StatusValidating, StatusReady, StatusBlocked,
		StatusCompleting, StatusCompleted, StatusFailed, StatusAbandoned:
		return true
	default:
		return false
	}
}

// IsTerminal reports whether the session can no longer progress.
func (s SessionStatus) IsTerminal() bool {
	switch s {
	case StatusCompleted, StatusFailed, StatusAbandoned:
		return true
	default:
		return false
	}
}

// CanPatch reports whether preferences may be updated.
func (s SessionStatus) CanPatch() bool {
	switch s {
	case StatusStarted, StatusValidating, StatusReady, StatusBlocked:
		return true
	default:
		return false
	}
}

// DeliveryOption is the chosen fulfillment mode.
type DeliveryOption string

const (
	DeliveryInstant    DeliveryOption = "instant"
	DeliveryScheduled  DeliveryOption = "scheduled"
	DeliveryPriority   DeliveryOption = "priority"
	DeliveryEconomy    DeliveryOption = "economy"
	DeliveryPickup     DeliveryOption = "pickup"
	DeliveryCorporate  DeliveryOption = "corporate"
)

// Valid reports whether the delivery option is recognized.
func (d DeliveryOption) Valid() bool {
	switch d {
	case DeliveryInstant, DeliveryScheduled, DeliveryPriority,
		DeliveryEconomy, DeliveryPickup, DeliveryCorporate:
		return true
	default:
		return false
	}
}

// AddressSnapshot is a denormalized delivery address at checkout time.
type AddressSnapshot struct {
	Label       string  `json:"label,omitempty"`
	Line1       string  `json:"line1,omitempty"`
	Line2       string  `json:"line2,omitempty"`
	City        string  `json:"city,omitempty"`
	District    string  `json:"district,omitempty"`
	PostalCode  string  `json:"postalCode,omitempty"`
	Country     string  `json:"country,omitempty"`
	Lat         float64 `json:"lat,omitempty"`
	Lng         float64 `json:"lng,omitempty"`
	ContactName string  `json:"contactName,omitempty"`
	Phone       string  `json:"phone,omitempty"`
}

// SlotSnapshot is a scheduled delivery window.
type SlotSnapshot struct {
	SlotID    string     `json:"slotId,omitempty"`
	StartsAt  *time.Time `json:"startsAt,omitempty"`
	EndsAt    *time.Time `json:"endsAt,omitempty"`
	Warehouse string     `json:"warehouse,omitempty"`
}

// GiftPrefs captures gift wrapping / message preferences.
type GiftPrefs struct {
	Enabled bool   `json:"enabled"`
	Message string `json:"message,omitempty"`
	From    string `json:"from,omitempty"`
}

// InvoicePrefs captures e-invoice / corporate billing prefs.
type InvoicePrefs struct {
	WantInvoice bool   `json:"wantInvoice"`
	TaxID       string `json:"taxId,omitempty"`
	CompanyName string `json:"companyName,omitempty"`
	Address     string `json:"address,omitempty"`
}

// SubstitutionPolicy controls out-of-stock replacements.
type SubstitutionPolicy string

const (
	SubstitutionAllow   SubstitutionPolicy = "allow"
	SubstitutionAsk     SubstitutionPolicy = "ask"
	SubstitutionRefuse  SubstitutionPolicy = "refuse"
)

// Valid reports whether the substitution policy is recognized.
func (p SubstitutionPolicy) Valid() bool {
	switch p {
	case SubstitutionAllow, SubstitutionAsk, SubstitutionRefuse, "":
		return true
	default:
		return false
	}
}

// QuoteSnapshot stores the last pricing preview (minor units).
type QuoteSnapshot struct {
	QuoteID         string    `json:"quoteId,omitempty"`
	Currency        string    `json:"currency,omitempty"`
	SubtotalMinor   int64     `json:"subtotalMinor"`
	DiscountMinor   int64     `json:"discountMinor"`
	TaxMinor        int64     `json:"taxMinor"`
	DeliveryMinor   int64     `json:"deliveryMinor"`
	ServiceMinor    int64     `json:"serviceMinor"`
	PackagingMinor  int64     `json:"packagingMinor"`
	TipMinor        int64     `json:"tipMinor"`
	TotalMinor      int64     `json:"totalMinor"`
	QuotedAt        time.Time `json:"quotedAt,omitempty"`
	LineCount       int       `json:"lineCount,omitempty"`
}

// ValidationResults is the persisted outcome of the validation pipeline.
type ValidationResults struct {
	Passed   bool              `json:"passed"`
	Issues   []ValidationIssue `json:"issues,omitempty"`
	CheckedAt time.Time        `json:"checkedAt,omitempty"`
}

// Session is the checkout aggregate.
type Session struct {
	ID              uuid.UUID
	TenantID        uuid.UUID
	CartID          uuid.UUID
	PrincipalID     uuid.UUID
	Status          SessionStatus
	DeliveryOption  DeliveryOption
	Address         AddressSnapshot
	Slot            SlotSnapshot
	Gift            GiftPrefs
	Invoice         InvoicePrefs
	Substitutions   SubstitutionPolicy
	Notes           string
	TipMinor        int64
	Currency        string
	Validation      ValidationResults
	Quote           QuoteSnapshot
	OrderID         string // opaque order-service id
	IdempotencyKey  string
	RecoveryToken   string
	CityID          string
	CouponCodes     []string
	Version         int64
	Metadata        map[string]any
	CreatedAt       time.Time
	UpdatedAt       time.Time
	CompletedAt     *time.Time
	AbandonedAt     *time.Time
	FailedAt        *time.Time
}

// sessionTransitions encodes the checkout state machine (ARCHITECTURE.md).
var sessionTransitions = map[SessionStatus][]SessionStatus{
	StatusStarted: {
		StatusValidating,
		StatusAbandoned,
		StatusFailed,
	},
	StatusValidating: {
		StatusReady,
		StatusBlocked,
		StatusAbandoned,
		StatusFailed,
	},
	StatusReady: {
		StatusValidating, // refresh / re-validate
		StatusCompleting,
		StatusAbandoned,
		StatusFailed,
	},
	StatusBlocked: {
		StatusValidating, // refresh after fix
		StatusAbandoned,
		StatusFailed,
	},
	StatusCompleting: {
		StatusCompleted,
		StatusFailed,
	},
	StatusCompleted: {},
	StatusFailed: {
		StatusAbandoned,
	},
	StatusAbandoned: {},
}

// ValidateTransition checks whether from→to is allowed.
func ValidateTransition(from, to SessionStatus) error {
	if !from.Valid() {
		return fmt.Errorf("%w: unknown status %q", ErrInvalidArgument, from)
	}
	if !to.Valid() {
		return fmt.Errorf("%w: unknown status %q", ErrInvalidArgument, to)
	}
	if from == to {
		return nil
	}
	allowed, ok := sessionTransitions[from]
	if !ok {
		return fmt.Errorf("%w: %s → %s", ErrInvalidTransition, from, to)
	}
	for _, a := range allowed {
		if a == to {
			return nil
		}
	}
	return fmt.Errorf("%w: %s → %s", ErrInvalidTransition, from, to)
}

// Validate checks structural invariants of a session.
func (s Session) Validate() error {
	if s.ID == uuid.Nil {
		return fmt.Errorf("%w: session id required", ErrInvalidArgument)
	}
	if s.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if s.CartID == uuid.Nil {
		return fmt.Errorf("%w: cart_id required", ErrInvalidArgument)
	}
	if s.PrincipalID == uuid.Nil {
		return fmt.Errorf("%w: principal_id required", ErrInvalidArgument)
	}
	if !s.Status.Valid() {
		return fmt.Errorf("%w: invalid status %q", ErrInvalidArgument, s.Status)
	}
	if s.DeliveryOption != "" && !s.DeliveryOption.Valid() {
		return fmt.Errorf("%w: invalid delivery option %q", ErrInvalidArgument, s.DeliveryOption)
	}
	if !s.Substitutions.Valid() {
		return fmt.Errorf("%w: invalid substitutions %q", ErrInvalidArgument, s.Substitutions)
	}
	if s.TipMinor < 0 {
		return fmt.Errorf("%w: tip", ErrNegativeMoney)
	}
	if s.Currency != "" {
		if _, err := NewMoney(0, s.Currency); err != nil {
			return err
		}
	}
	if strings.TrimSpace(s.IdempotencyKey) == "" {
		return fmt.Errorf("%w: idempotency_key required", ErrInvalidArgument)
	}
	if s.Version < 1 {
		return fmt.Errorf("%w: version must be >= 1", ErrInvalidArgument)
	}
	return nil
}
