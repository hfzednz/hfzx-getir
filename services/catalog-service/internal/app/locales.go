package app

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/catalog-service/internal/domain"
)

// UpsertProductLocaleInput upserts localized content.
type UpsertProductLocaleInput struct {
	TenantID         uuid.UUID
	ProductID        uuid.UUID
	Lang             string
	Title            string
	Subtitle         string
	Description      string
	ShortDescription string
}

// UpsertProductLocale upserts localized product content.
func (d *Deps) UpsertProductLocale(ctx context.Context, in UpsertProductLocaleInput) (domain.ProductLocale, error) {
	p, err := d.getProduct(ctx, in.TenantID, in.ProductID)
	if err != nil {
		return domain.ProductLocale{}, err
	}
	if err := d.ensureEditable(p); err != nil {
		return domain.ProductLocale{}, err
	}
	now := d.now()
	existing, err := d.Locales.Get(ctx, in.TenantID, in.ProductID, in.Lang)
	if err != nil && err != domain.ErrNotFound {
		return domain.ProductLocale{}, err
	}
	l := domain.ProductLocale{
		ID:               d.newID(),
		ProductID:        in.ProductID,
		TenantID:         in.TenantID,
		Lang:             strings.TrimSpace(in.Lang),
		Title:            strings.TrimSpace(in.Title),
		Subtitle:         strings.TrimSpace(in.Subtitle),
		Description:      strings.TrimSpace(in.Description),
		ShortDescription: strings.TrimSpace(in.ShortDescription),
		Metadata:         map[string]any{},
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err == nil {
		l.ID = existing.ID
		l.CreatedAt = existing.CreatedAt
	}
	if err := l.Validate(); err != nil {
		return domain.ProductLocale{}, err
	}
	if err := d.Locales.Upsert(ctx, l); err != nil {
		return domain.ProductLocale{}, err
	}
	d.indexProduct(ctx, in.TenantID, in.ProductID)
	return l, nil
}

// GetProductLocale returns localized content.
func (d *Deps) GetProductLocale(ctx context.Context, tenantID, productID uuid.UUID, lang string) (domain.ProductLocale, error) {
	return d.Locales.Get(ctx, tenantID, productID, lang)
}

// ListProductLocales lists all locales for a product.
func (d *Deps) ListProductLocales(ctx context.Context, tenantID, productID uuid.UUID) ([]domain.ProductLocale, error) {
	return d.Locales.ListByProduct(ctx, tenantID, productID)
}
