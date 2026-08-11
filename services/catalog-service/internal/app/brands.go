package app

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/catalog-service/internal/domain"
)

// CreateBrandInput creates a brand.
type CreateBrandInput struct {
	TenantID    uuid.UUID
	Name        string
	Slug        string
	Description string
	LogoURL     string
	WebsiteURL  string
	CountryCode string
}

// CreateBrand inserts a brand master record.
func (d *Deps) CreateBrand(ctx context.Context, in CreateBrandInput) (domain.Brand, error) {
	now := d.now()
	b := domain.Brand{
		ID:          d.newID(),
		TenantID:    in.TenantID,
		Name:        strings.TrimSpace(in.Name),
		Slug:        strings.TrimSpace(in.Slug),
		Description: strings.TrimSpace(in.Description),
		LogoURL:     strings.TrimSpace(in.LogoURL),
		WebsiteURL:  strings.TrimSpace(in.WebsiteURL),
		CountryCode: strings.ToUpper(strings.TrimSpace(in.CountryCode)),
		Metadata:    map[string]any{},
		Status:      domain.BrandStatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := b.Validate(); err != nil {
		return domain.Brand{}, err
	}
	if _, err := d.Brands.GetBySlug(ctx, in.TenantID, b.Slug); err == nil {
		return domain.Brand{}, domain.ErrAlreadyExists
	}
	if err := d.Brands.Create(ctx, b); err != nil {
		return domain.Brand{}, err
	}
	d.publishEvent(ctx, domain.EventBrandChanged, in.TenantID, uuid.Nil, map[string]any{"brandId": b.ID, "action": "created"})
	return b, nil
}

// GetBrand returns a brand by id.
func (d *Deps) GetBrand(ctx context.Context, tenantID, brandID uuid.UUID) (domain.Brand, error) {
	return d.Brands.GetByID(ctx, tenantID, brandID)
}

// ListBrands lists brands for a tenant.
func (d *Deps) ListBrands(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.Brand, int, error) {
	if limit <= 0 {
		limit = 50
	}
	return d.Brands.List(ctx, tenantID, limit, offset)
}
