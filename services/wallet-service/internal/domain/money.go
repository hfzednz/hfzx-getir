package domain

import (
	"fmt"
	"strings"
)

// Money is an integer minor-units amount with an ISO-4217 currency.
type Money struct {
	AmountMinor int64
	Currency    string
}

// NewMoney constructs Money after validating currency and non-negativity.
func NewMoney(amountMinor int64, currency string) (Money, error) {
	m := Money{AmountMinor: amountMinor, Currency: strings.ToUpper(strings.TrimSpace(currency))}
	if err := m.Validate(); err != nil {
		return Money{}, err
	}
	return m, nil
}

// Validate checks currency shape and non-negative amount.
func (m Money) Validate() error {
	if len(m.Currency) != 3 {
		return fmt.Errorf("%w: currency must be ISO-4217 (got %q)", ErrInvalidArgument, m.Currency)
	}
	for _, c := range m.Currency {
		if c < 'A' || c > 'Z' {
			return fmt.Errorf("%w: currency must be uppercase letters (got %q)", ErrInvalidArgument, m.Currency)
		}
	}
	if m.AmountMinor < 0 {
		return fmt.Errorf("%w: %d", ErrNegativeMoney, m.AmountMinor)
	}
	return nil
}

// MustMoney panics on invalid money — for tests/static fixtures only.
func MustMoney(amountMinor int64, currency string) Money {
	m, err := NewMoney(amountMinor, currency)
	if err != nil {
		panic(err)
	}
	return m
}
