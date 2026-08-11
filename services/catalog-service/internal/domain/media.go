package domain

import (
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// MediaKind classifies product media references.
type MediaKind string

const (
	MediaKindImage MediaKind = "image"
	MediaKindVideo MediaKind = "video"
	MediaKind360   MediaKind = "360"
	MediaKindAR    MediaKind = "ar"
	MediaKindPDF   MediaKind = "pdf"
	MediaKindAudio MediaKind = "audio"
)

func (k MediaKind) Valid() bool {
	switch k {
	case MediaKindImage, MediaKindVideo, MediaKind360, MediaKindAR, MediaKindPDF, MediaKindAudio:
		return true
	default:
		return false
	}
}

const maxAltTextLen = 500

// ProductMedia is a reference to a media-service asset (+ CDN URL snapshot).
type ProductMedia struct {
	ID           uuid.UUID
	ProductID    uuid.UUID
	VariantID    *uuid.UUID
	TenantID     uuid.UUID
	MediaAssetID uuid.UUID
	Kind         MediaKind
	SortOrder    int
	CDNURL       string
	AltText      string
	Locale       string
	IsPrimary    bool
	Metadata     map[string]any
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Validate checks structural invariants.
func (m ProductMedia) Validate() error {
	if m.ID == uuid.Nil {
		return fmt.Errorf("%w: product_media id required", ErrInvalidArgument)
	}
	if m.ProductID == uuid.Nil {
		return fmt.Errorf("%w: product_id required", ErrInvalidArgument)
	}
	if m.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if m.MediaAssetID == uuid.Nil {
		return fmt.Errorf("%w: media_asset_id required", ErrInvalidArgument)
	}
	if !m.Kind.Valid() {
		return fmt.Errorf("%w: invalid media kind %q", ErrInvalidArgument, m.Kind)
	}
	if utf8.RuneCountInString(m.CDNURL) > maxURLLen {
		return fmt.Errorf("%w: cdn_url too long", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(m.AltText) > maxAltTextLen {
		return fmt.Errorf("%w: alt_text too long", ErrInvalidArgument)
	}
	if m.Locale != "" {
		if err := ValidateLang(m.Locale); err != nil {
			return err
		}
	}
	if m.VariantID != nil && *m.VariantID == uuid.Nil {
		return fmt.Errorf("%w: variant_id must not be nil UUID", ErrInvalidArgument)
	}
	return nil
}
