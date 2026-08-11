package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AccountType classifies wallet sub-accounts.
type AccountType string

const (
	AccountCash     AccountType = "cash"
	AccountRefund   AccountType = "refund"
	AccountPromo    AccountType = "promo"
	AccountCashback AccountType = "cashback"
	AccountGift     AccountType = "gift"
)

// Valid reports whether the account type is recognized.
func (t AccountType) Valid() bool {
	switch t {
	case AccountCash, AccountRefund, AccountPromo, AccountCashback, AccountGift:
		return true
	default:
		return false
	}
}

// AllAccountTypes returns the standard account set per wallet.
func AllAccountTypes() []AccountType {
	return []AccountType{AccountCash, AccountRefund, AccountPromo, AccountCashback, AccountGift}
}

// Wallet is the aggregate root for a principal's balances.
type Wallet struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	PrincipalID uuid.UUID
	Currency    string
	Active      bool
	Version     int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Validate checks wallet invariants.
func (w Wallet) Validate() error {
	if w.ID == uuid.Nil {
		return fmt.Errorf("%w: wallet id required", ErrInvalidArgument)
	}
	if w.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if w.PrincipalID == uuid.Nil {
		return fmt.Errorf("%w: principal_id required", ErrInvalidArgument)
	}
	if _, err := NewMoney(0, w.Currency); err != nil {
		return err
	}
	return nil
}

// Account is a sub-ledger within a wallet.
// Available = BalanceMinor - HeldMinor (never negative).
type Account struct {
	ID           uuid.UUID
	WalletID     uuid.UUID
	TenantID     uuid.UUID
	AccountType  AccountType
	BalanceMinor int64
	HeldMinor    int64
	Currency     string
	Version      int64
	UpdatedAt    time.Time
}

// Available returns spendable balance (balance - held).
func (a Account) Available() int64 {
	return a.BalanceMinor - a.HeldMinor
}

// Validate checks account invariants.
func (a Account) Validate() error {
	if a.ID == uuid.Nil {
		return fmt.Errorf("%w: account id required", ErrInvalidArgument)
	}
	if a.WalletID == uuid.Nil {
		return fmt.Errorf("%w: wallet_id required", ErrInvalidArgument)
	}
	if !a.AccountType.Valid() {
		return fmt.Errorf("%w: invalid account type %q", ErrInvalidArgument, a.AccountType)
	}
	if a.BalanceMinor < 0 || a.HeldMinor < 0 {
		return fmt.Errorf("%w: balance/held must be non-negative", ErrNegativeMoney)
	}
	if a.Available() < 0 {
		return fmt.Errorf("%w: available would be negative", ErrOverdraft)
	}
	return nil
}

// EntryKind classifies ledger entries.
type EntryKind string

const (
	EntryCredit  EntryKind = "credit"
	EntryDebit   EntryKind = "debit"
	EntryHold    EntryKind = "hold"
	EntryRelease EntryKind = "release"
	EntryAdjust  EntryKind = "adjust"
)

// Entry is an append-only wallet ledger row.
type Entry struct {
	ID             uuid.UUID
	WalletID       uuid.UUID
	AccountID      uuid.UUID
	TenantID       uuid.UUID
	Kind           EntryKind
	AmountMinor    int64
	Currency       string
	BalanceAfter   int64
	HeldAfter      int64
	Reference      string
	IdempotencyKey string
	Metadata       map[string]any
	CreatedAt      time.Time
}

// HoldStatus is the hold lifecycle.
type HoldStatus string

const (
	HoldActive   HoldStatus = "active"
	HoldReleased HoldStatus = "released"
	HoldCaptured HoldStatus = "captured" // released via debit
)

// Hold reserves available balance.
type Hold struct {
	ID             uuid.UUID
	WalletID       uuid.UUID
	AccountID      uuid.UUID
	TenantID       uuid.UUID
	AmountMinor    int64
	Currency       string
	Status         HoldStatus
	Reference      string
	IdempotencyKey string
	CreatedAt      time.Time
	ReleasedAt     *time.Time
}

// Transfer moves funds between accounts (same wallet) or wallets (same tenant).
type Transfer struct {
	ID               uuid.UUID
	TenantID         uuid.UUID
	FromWalletID     uuid.UUID
	FromAccountID    uuid.UUID
	ToWalletID       uuid.UUID
	ToAccountID      uuid.UUID
	AmountMinor      int64
	Currency         string
	IdempotencyKey   string
	Reference        string
	CreatedAt        time.Time
}

// Limit caps daily/periodic wallet activity.
type Limit struct {
	ID           uuid.UUID
	WalletID     uuid.UUID
	TenantID     uuid.UUID
	LimitType    string // daily_debit | daily_credit | max_balance
	AmountMinor  int64
	Currency     string
	WindowKey    string // e.g. 2026-08-06 for daily
	UsedMinor    int64
	UpdatedAt    time.Time
}
