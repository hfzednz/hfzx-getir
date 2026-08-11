package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// AccountType classifies chart-of-accounts entries.
type AccountType string

const (
	AccountTypeAsset     AccountType = "asset"
	AccountTypeLiability AccountType = "liability"
	AccountTypeRevenue   AccountType = "revenue"
	AccountTypeExpense   AccountType = "expense"
	AccountTypeClearing  AccountType = "clearing"
)

// Valid reports whether the account type is recognized.
func (t AccountType) Valid() bool {
	switch t {
	case AccountTypeAsset, AccountTypeLiability, AccountTypeRevenue, AccountTypeExpense, AccountTypeClearing:
		return true
	default:
		return false
	}
}

// Account is a chart-of-accounts entry.
type Account struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Code        string
	Name        string
	Type        AccountType
	Currency    string
	Active      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Version     int64
}

// Validate checks account invariants.
func (a Account) Validate() error {
	if a.ID == uuid.Nil {
		return fmt.Errorf("%w: account id required", ErrInvalidArgument)
	}
	if a.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	code := strings.TrimSpace(a.Code)
	if code == "" {
		return fmt.Errorf("%w: account code required", ErrInvalidArgument)
	}
	if strings.TrimSpace(a.Name) == "" {
		return fmt.Errorf("%w: account name required", ErrInvalidArgument)
	}
	if !a.Type.Valid() {
		return fmt.Errorf("%w: invalid account type %q", ErrInvalidArgument, a.Type)
	}
	if _, err := NewMoney(0, a.Currency); err != nil {
		return err
	}
	return nil
}
