package domain

import (
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// SEOEntityType identifies the catalog entity SEO is attached to.
type SEOEntityType string

const (
	SEOEntityProduct    SEOEntityType = "product"
	SEOEntityCategory   SEOEntityType = "category"
	SEOEntityBrand      SEOEntityType = "brand"
	SEOEntityCollection SEOEntityType = "collection"
)

func (t SEOEntityType) Valid() bool {
	switch t {
	case SEOEntityProduct, SEOEntityCategory, SEOEntityBrand, SEOEntityCollection:
		return true
	default:
		return false
	}
}

const (
	maxMetaTitleLen = 200
	maxMetaDescLen  = 500
)

// SEO holds search / social metadata for a catalog entity.
type SEO struct {
	ID                 uuid.UUID
	TenantID           uuid.UUID
	EntityType         SEOEntityType
	EntityID           uuid.UUID
	Lang               string
	Slug               string
	MetaTitle          string
	MetaDescription    string
	MetaKeywords       string
	CanonicalURL       string
	OGTitle            string
	OGDescription      string
	OGImageURL         string
	TwitterCard        string
	TwitterTitle       string
	TwitterDescription string
	TwitterImageURL    string
	JSONLD             map[string]any
	Robots             string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// Validate checks structural invariants.
func (s SEO) Validate() error {
	if s.ID == uuid.Nil {
		return fmt.Errorf("%w: seo id required", ErrInvalidArgument)
	}
	if s.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if !s.EntityType.Valid() {
		return fmt.Errorf("%w: invalid seo entity type %q", ErrInvalidArgument, s.EntityType)
	}
	if s.EntityID == uuid.Nil {
		return fmt.Errorf("%w: entity_id required", ErrInvalidArgument)
	}
	if err := ValidateLang(s.Lang); err != nil {
		return err
	}
	if s.Slug != "" {
		if err := ValidateSlug(s.Slug); err != nil {
			return err
		}
	}
	if utf8.RuneCountInString(s.MetaTitle) > maxMetaTitleLen {
		return fmt.Errorf("%w: meta_title too long", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(s.MetaDescription) > maxMetaDescLen {
		return fmt.Errorf("%w: meta_description too long", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(s.CanonicalURL) > maxURLLen {
		return fmt.Errorf("%w: canonical_url too long", ErrInvalidArgument)
	}
	if s.JSONLD == nil {
		return fmt.Errorf("%w: jsonld required", ErrInvalidArgument)
	}
	return nil
}
