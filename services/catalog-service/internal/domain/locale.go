package domain

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

var langRegexp = regexp.MustCompile(`^[a-z]{2}(-[A-Z]{2})?$`)

const (
	maxTitleLen       = 500
	maxSubtitleLen    = 500
	maxDescriptionLen = 20000
	maxLocaleFieldLen = 10000
)

// ValidateLang checks a BCP-47 language tag (en, tr, en-US).
func ValidateLang(lang string) error {
	lang = strings.TrimSpace(lang)
	if lang == "" {
		return fmt.Errorf("%w: lang required", ErrInvalidArgument)
	}
	if !langRegexp.MatchString(lang) {
		return fmt.Errorf("%w: invalid lang %q", ErrInvalidArgument, lang)
	}
	return nil
}

// ProductLocale holds localized product content.
type ProductLocale struct {
	ID               uuid.UUID
	ProductID        uuid.UUID
	TenantID         uuid.UUID
	Lang             string
	Title            string
	Subtitle         string
	Description      string
	ShortDescription string
	Specs            string
	Usage            string
	Warnings         string
	Ingredients      string
	Allergens        string
	Nutrition        string
	Storage          string
	Origin           string
	Metadata         map[string]any
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// Validate checks structural invariants.
func (l ProductLocale) Validate() error {
	if l.ID == uuid.Nil {
		return fmt.Errorf("%w: product_locale id required", ErrInvalidArgument)
	}
	if l.ProductID == uuid.Nil {
		return fmt.Errorf("%w: product_id required", ErrInvalidArgument)
	}
	if l.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if err := ValidateLang(l.Lang); err != nil {
		return err
	}
	if strings.TrimSpace(l.Title) == "" {
		return fmt.Errorf("%w: title required", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(l.Title) > maxTitleLen {
		return fmt.Errorf("%w: title too long", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(l.Subtitle) > maxSubtitleLen {
		return fmt.Errorf("%w: subtitle too long", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(l.Description) > maxDescriptionLen {
		return fmt.Errorf("%w: description too long", ErrInvalidArgument)
	}
	for _, field := range []struct {
		name  string
		value string
	}{
		{"short_description", l.ShortDescription},
		{"specs", l.Specs},
		{"usage", l.Usage},
		{"warnings", l.Warnings},
		{"ingredients", l.Ingredients},
		{"allergens", l.Allergens},
		{"nutrition", l.Nutrition},
		{"storage", l.Storage},
		{"origin", l.Origin},
	} {
		if utf8.RuneCountInString(field.value) > maxLocaleFieldLen {
			return fmt.Errorf("%w: %s too long", ErrInvalidArgument, field.name)
		}
	}
	return nil
}
