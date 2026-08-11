package domain

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// SKUIdentifierType classifies barcode / code schemes.
type SKUIdentifierType string

const (
	SKUTypeBarcode  SKUIdentifierType = "barcode"
	SKUTypeQR       SKUIdentifierType = "qr"
	SKUTypeEAN      SKUIdentifierType = "ean"
	SKUTypeUPC      SKUIdentifierType = "upc"
	SKUTypeGTIN     SKUIdentifierType = "gtin"
	SKUTypeInternal SKUIdentifierType = "internal"
	SKUTypeSupplier SKUIdentifierType = "supplier"
	SKUTypeWarehouse SKUIdentifierType = "warehouse"
)

func (t SKUIdentifierType) Valid() bool {
	switch t {
	case SKUTypeBarcode, SKUTypeQR, SKUTypeEAN, SKUTypeUPC, SKUTypeGTIN,
		SKUTypeInternal, SKUTypeSupplier, SKUTypeWarehouse:
		return true
	default:
		return false
	}
}

var digitsOnly = regexp.MustCompile(`^\d+$`)

const maxSKUIdentifierLen = 128

// SKUIdentifier is an external or internal code bound to a variant.
type SKUIdentifier struct {
	ID        uuid.UUID
	VariantID uuid.UUID
	TenantID  uuid.UUID
	Type      SKUIdentifierType
	Value     string
	IsPrimary bool
	Metadata  map[string]any
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Validate checks structural invariants and type-specific formats.
func (s SKUIdentifier) Validate() error {
	if s.ID == uuid.Nil {
		return fmt.Errorf("%w: sku identifier id required", ErrInvalidArgument)
	}
	if s.VariantID == uuid.Nil {
		return fmt.Errorf("%w: variant_id required", ErrInvalidArgument)
	}
	if s.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if !s.Type.Valid() {
		return fmt.Errorf("%w: invalid sku identifier type %q", ErrInvalidArgument, s.Type)
	}
	value := strings.TrimSpace(s.Value)
	if value == "" {
		return fmt.Errorf("%w: value required", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(value) > maxSKUIdentifierLen {
		return fmt.Errorf("%w: value too long", ErrInvalidArgument)
	}
	if err := ValidateBarcodeFormat(s.Type, value); err != nil {
		return err
	}
	return nil
}

// ValidateBarcodeFormat validates identifier value shape by type.
// EAN-8/13, UPC-A, GTIN-14 use GS1 check digits; others are length/charset checks.
func ValidateBarcodeFormat(t SKUIdentifierType, value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("%w: empty value", ErrInvalidBarcode)
	}

	switch t {
	case SKUTypeEAN:
		if !digitsOnly.MatchString(value) {
			return fmt.Errorf("%w: EAN must be digits", ErrInvalidBarcode)
		}
		switch len(value) {
		case 8, 13:
			if !gs1CheckDigitValid(value) {
				return fmt.Errorf("%w: EAN check digit mismatch", ErrInvalidBarcode)
			}
		default:
			return fmt.Errorf("%w: EAN must be 8 or 13 digits", ErrInvalidBarcode)
		}
	case SKUTypeUPC:
		if !digitsOnly.MatchString(value) || len(value) != 12 {
			return fmt.Errorf("%w: UPC-A must be 12 digits", ErrInvalidBarcode)
		}
		if !gs1CheckDigitValid(value) {
			return fmt.Errorf("%w: UPC check digit mismatch", ErrInvalidBarcode)
		}
	case SKUTypeGTIN:
		if !digitsOnly.MatchString(value) {
			return fmt.Errorf("%w: GTIN must be digits", ErrInvalidBarcode)
		}
		switch len(value) {
		case 8, 12, 13, 14:
			if !gs1CheckDigitValid(value) {
				return fmt.Errorf("%w: GTIN check digit mismatch", ErrInvalidBarcode)
			}
		default:
			return fmt.Errorf("%w: GTIN must be 8, 12, 13, or 14 digits", ErrInvalidBarcode)
		}
	case SKUTypeBarcode:
		if len(value) < 4 || len(value) > 48 {
			return fmt.Errorf("%w: barcode length out of range", ErrInvalidBarcode)
		}
	case SKUTypeQR, SKUTypeInternal, SKUTypeSupplier, SKUTypeWarehouse:
		if len(value) < 1 || len(value) > maxSKUIdentifierLen {
			return fmt.Errorf("%w: identifier length out of range", ErrInvalidBarcode)
		}
	default:
		return fmt.Errorf("%w: unknown type %q", ErrInvalidBarcode, t)
	}
	return nil
}

// gs1CheckDigitValid verifies the GS1 mod-10 check digit for the full code.
func gs1CheckDigitValid(code string) bool {
	if len(code) < 2 {
		return false
	}
	body := code[:len(code)-1]
	expected, err := strconv.Atoi(string(code[len(code)-1]))
	if err != nil {
		return false
	}
	return gs1CheckDigit(body) == expected
}

// gs1CheckDigit computes the GS1 check digit for a digit string without the check digit.
func gs1CheckDigit(body string) int {
	sum := 0
	// From the right: odd positions ×3, even positions ×1 (GS1).
	for i, r := range body {
		d := int(r - '0')
		posFromRight := len(body) - i
		if posFromRight%2 == 1 {
			sum += d * 3
		} else {
			sum += d
		}
	}
	return (10 - (sum % 10)) % 10
}

// NormalizeBarcode strips whitespace and uppercases non-digit identifiers.
func NormalizeBarcode(t SKUIdentifierType, value string) string {
	value = strings.TrimSpace(value)
	switch t {
	case SKUTypeEAN, SKUTypeUPC, SKUTypeGTIN, SKUTypeBarcode:
		return strings.Map(func(r rune) rune {
			if r >= '0' && r <= '9' {
				return r
			}
			if r == ' ' || r == '-' {
				return -1
			}
			return r
		}, value)
	default:
		return strings.ToUpper(value)
	}
}
