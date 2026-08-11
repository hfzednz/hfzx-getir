package domain

import (
	"fmt"
	"strings"
)

// Money is an integer minor-units amount with an ISO-4217 currency.
// Never use float for monetary values.
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

// ZeroMoney returns a zero amount in the given currency.
func ZeroMoney(currency string) (Money, error) {
	return NewMoney(0, currency)
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

// IsZero reports whether the amount is zero.
func (m Money) IsZero() bool {
	return m.AmountMinor == 0
}

// SameCurrency reports whether both moneys share a currency.
func (m Money) SameCurrency(other Money) bool {
	return m.Currency == other.Currency
}

// Add returns m + other. Currencies must match.
func (m Money) Add(other Money) (Money, error) {
	if !m.SameCurrency(other) {
		return Money{}, fmt.Errorf("%w: %s vs %s", ErrCurrencyMismatch, m.Currency, other.Currency)
	}
	sum := m.AmountMinor + other.AmountMinor
	if sum < 0 {
		return Money{}, fmt.Errorf("%w: overflow/underflow", ErrInvariant)
	}
	return Money{AmountMinor: sum, Currency: m.Currency}, nil
}

// Sub returns m - other. Result must stay non-negative.
func (m Money) Sub(other Money) (Money, error) {
	if !m.SameCurrency(other) {
		return Money{}, fmt.Errorf("%w: %s vs %s", ErrCurrencyMismatch, m.Currency, other.Currency)
	}
	if other.AmountMinor > m.AmountMinor {
		return Money{}, fmt.Errorf("%w: insufficient amount %d < %d", ErrInvariant, m.AmountMinor, other.AmountMinor)
	}
	return Money{AmountMinor: m.AmountMinor - other.AmountMinor, Currency: m.Currency}, nil
}

// MulQty multiplies unit price by a positive quantity (integer only).
func (m Money) MulQty(qty int) (Money, error) {
	if qty <= 0 {
		return Money{}, fmt.Errorf("%w: qty must be > 0", ErrInvalidArgument)
	}
	product := m.AmountMinor * int64(qty)
	if m.AmountMinor > 0 && product/m.AmountMinor != int64(qty) {
		return Money{}, fmt.Errorf("%w: mul overflow", ErrInvariant)
	}
	if product < 0 {
		return Money{}, fmt.Errorf("%w: mul overflow", ErrInvariant)
	}
	return Money{AmountMinor: product, Currency: m.Currency}, nil
}

// MustMoney panics on invalid money — for tests/static fixtures only.
func MustMoney(amountMinor int64, currency string) Money {
	m, err := NewMoney(amountMinor, currency)
	if err != nil {
		panic(err)
	}
	return m
}
