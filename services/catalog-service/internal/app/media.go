package app

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/catalog-service/internal/domain"
)

// AttachMediaInput attaches a media reference to a product.
type AttachMediaInput struct {
	TenantID     uuid.UUID
	ProductID    uuid.UUID
	VariantID    *uuid.UUID
	MediaAssetID uuid.UUID
	Kind         domain.MediaKind
	SortOrder    int
	AltText      string
	Locale       string
	IsPrimary    bool
}

// AttachMedia stores a media reference, resolving CDN URL via media-service port.
func (d *Deps) AttachMedia(ctx context.Context, in AttachMediaInput) (domain.ProductMedia, error) {
	if _, err := d.getProduct(ctx, in.TenantID, in.ProductID); err != nil {
		return domain.ProductMedia{}, err
	}
	cdnURL := ""
	if d.MediaClient != nil {
		asset, err := d.MediaClient.GetAsset(ctx, in.TenantID, in.MediaAssetID)
		if err == nil {
			cdnURL = asset.CDNURL
			if in.Kind == "" {
				in.Kind = asset.Kind
			}
			if in.AltText == "" {
				in.AltText = asset.AltText
			}
		}
	}
	if in.Kind == "" {
		in.Kind = domain.MediaKindImage
	}
	now := d.now()
	m := domain.ProductMedia{
		ID:           d.newID(),
		ProductID:    in.ProductID,
		VariantID:    in.VariantID,
		TenantID:     in.TenantID,
		MediaAssetID: in.MediaAssetID,
		Kind:         in.Kind,
		SortOrder:    in.SortOrder,
		CDNURL:       strings.TrimSpace(cdnURL),
		AltText:      strings.TrimSpace(in.AltText),
		Locale:       strings.TrimSpace(in.Locale),
		IsPrimary:    in.IsPrimary,
		Metadata:     map[string]any{},
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := m.Validate(); err != nil {
		return domain.ProductMedia{}, err
	}
	if err := d.Media.Create(ctx, m); err != nil {
		return domain.ProductMedia{}, err
	}
	d.publishEvent(ctx, domain.EventMediaAttached, in.TenantID, in.ProductID, map[string]any{"mediaId": m.ID})
	return m, nil
}

// ListProductMedia lists media refs for a product.
func (d *Deps) ListProductMedia(ctx context.Context, tenantID, productID uuid.UUID) ([]domain.ProductMedia, error) {
	return d.Media.ListByProduct(ctx, tenantID, productID)
}

// DetachMedia removes a media reference.
func (d *Deps) DetachMedia(ctx context.Context, tenantID, mediaID uuid.UUID) error {
	return d.Media.Delete(ctx, tenantID, mediaID)
}
