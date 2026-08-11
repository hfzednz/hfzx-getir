package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// JournalStatus is the journal lifecycle.
type JournalStatus string

const (
	JournalStatusDraft  JournalStatus = "draft"
	JournalStatusPosted JournalStatus = "posted"
)

// Valid reports whether the journal status is recognized.
func (s JournalStatus) Valid() bool {
	switch s {
	case JournalStatusDraft, JournalStatusPosted:
		return true
	default:
		return false
	}
}

// JournalLine is a single debit or credit entry (minor units).
// Exactly one of DebitMinor / CreditMinor must be > 0.
type JournalLine struct {
	ID          uuid.UUID
	AccountID   uuid.UUID
	AccountCode string
	DebitMinor  int64
	CreditMinor int64
	Currency    string
	Memo        string
}

// Validate checks line invariants.
func (l JournalLine) Validate() error {
	if l.AccountID == uuid.Nil {
		return fmt.Errorf("%w: account_id required", ErrInvalidArgument)
	}
	if l.DebitMinor < 0 || l.CreditMinor < 0 {
		return fmt.Errorf("%w: line amounts must be non-negative", ErrNegativeMoney)
	}
	if l.DebitMinor == 0 && l.CreditMinor == 0 {
		return fmt.Errorf("%w: line must have debit or credit", ErrInvalidArgument)
	}
	if l.DebitMinor > 0 && l.CreditMinor > 0 {
		return fmt.Errorf("%w: line cannot have both debit and credit", ErrInvalidArgument)
	}
	if _, err := NewMoney(0, l.Currency); err != nil {
		return err
	}
	return nil
}

// Journal is a double-entry journal header. Immutable after post.
type Journal struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	Status         JournalStatus
	Currency       string
	Reference      string // opaque external ref (payment/wallet/settlement id)
	Description    string
	IdempotencyKey string
	Lines          []JournalLine
	PostedAt       *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Version        int64
}

// Validate checks journal structural invariants (does not require balance for draft).
func (j Journal) Validate() error {
	if j.ID == uuid.Nil {
		return fmt.Errorf("%w: journal id required", ErrInvalidArgument)
	}
	if j.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if !j.Status.Valid() {
		return fmt.Errorf("%w: invalid journal status %q", ErrInvalidArgument, j.Status)
	}
	if _, err := NewMoney(0, j.Currency); err != nil {
		return err
	}
	if len(j.Lines) < 2 {
		return fmt.Errorf("%w: journal requires at least 2 lines", ErrInvalidArgument)
	}
	for i, line := range j.Lines {
		if err := line.Validate(); err != nil {
			return fmt.Errorf("%w: line[%d]", err, i)
		}
		if !strings.EqualFold(line.Currency, j.Currency) {
			return fmt.Errorf("%w: line[%d] currency", ErrCurrencyMismatch, i)
		}
	}
	return nil
}

// DebitTotal returns sum of debit minor units.
func (j Journal) DebitTotal() int64 {
	var sum int64
	for _, l := range j.Lines {
		sum += l.DebitMinor
	}
	return sum
}

// CreditTotal returns sum of credit minor units.
func (j Journal) CreditTotal() int64 {
	var sum int64
	for _, l := range j.Lines {
		sum += l.CreditMinor
	}
	return sum
}

// IsBalanced reports whether sum(debit) == sum(credit) and totals > 0.
func (j Journal) IsBalanced() bool {
	d, c := j.DebitTotal(), j.CreditTotal()
	return d == c && d > 0
}

// AssertBalanced returns ErrUnbalancedJournal when not balanced.
func (j Journal) AssertBalanced() error {
	if !j.IsBalanced() {
		return fmt.Errorf("%w: debit=%d credit=%d", ErrUnbalancedJournal, j.DebitTotal(), j.CreditTotal())
	}
	return nil
}

// AssertMutable returns ErrJournalImmutable when already posted.
func (j Journal) AssertMutable() error {
	if j.Status == JournalStatusPosted {
		return ErrJournalImmutable
	}
	return nil
}
