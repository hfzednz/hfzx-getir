package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// TaxRule defines a percentage tax in basis points (e.g. 1800 = 18.00%).
type TaxRule struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Code        string
	Name        string
	RateBps     int64 // basis points; 10000 = 100%
	Currency    string
	Active      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Validate checks tax rule invariants.
func (r TaxRule) Validate() error {
	if r.ID == uuid.Nil {
		return fmt.Errorf("%w: tax rule id required", ErrInvalidArgument)
	}
	if r.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if strings.TrimSpace(r.Code) == "" {
		return fmt.Errorf("%w: tax code required", ErrInvalidArgument)
	}
	if r.RateBps < 0 || r.RateBps > 10000 {
		return fmt.Errorf("%w: rate_bps must be 0..10000", ErrInvalidArgument)
	}
	if r.Currency != "" {
		if _, err := NewMoney(0, r.Currency); err != nil {
			return err
		}
	}
	return nil
}

// TaxResult is the output of TaxCalculate.
type TaxResult struct {
	BaseMinor int64
	TaxMinor  int64
	TotalMinor int64
	RateBps   int64
	TaxCode   string
	Currency  string
}

// CalculateTax applies rate_bps to baseMinor using integer arithmetic (round half up).
func CalculateTax(baseMinor int64, rateBps int64, currency, taxCode string) (TaxResult, error) {
	if baseMinor < 0 {
		return TaxResult{}, fmt.Errorf("%w: base", ErrNegativeMoney)
	}
	if rateBps < 0 || rateBps > 10000 {
		return TaxResult{}, fmt.Errorf("%w: rate_bps", ErrInvalidArgument)
	}
	if _, err := NewMoney(0, currency); err != nil {
		return TaxResult{}, err
	}
	// tax = round(base * rateBps / 10000)
	tax := (baseMinor*rateBps + 5000) / 10000
	return TaxResult{
		BaseMinor:  baseMinor,
		TaxMinor:   tax,
		TotalMinor: baseMinor + tax,
		RateBps:    rateBps,
		TaxCode:    taxCode,
		Currency:   strings.ToUpper(strings.TrimSpace(currency)),
	}, nil
}
