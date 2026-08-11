package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/nexora/catalog-service/internal/domain"
)

// UpsertComplianceInput upserts regulatory flags.
type UpsertComplianceInput struct {
	TenantID             uuid.UUID
	ProductID            uuid.UUID
	AgeRestriction       int
	IsHazardous          bool
	HazardClass          string
	IsPharmacy           bool
	RequiresPrescription bool
	IsFood               bool
	IsOrganic            bool
	IsHalal              bool
	IsVegan              bool
	IsGlutenFree         bool
	RestrictedCountries  []string
	AllowedCountries     []string
	Certificates         []map[string]any
}

// UpsertCompliance upserts product compliance metadata.
func (d *Deps) UpsertCompliance(ctx context.Context, in UpsertComplianceInput) (domain.ProductCompliance, error) {
	if _, err := d.getProduct(ctx, in.TenantID, in.ProductID); err != nil {
		return domain.ProductCompliance{}, err
	}
	if in.Certificates == nil {
		in.Certificates = []map[string]any{}
	}
	now := d.now()
	c := domain.ProductCompliance{
		ID:                   d.newID(),
		ProductID:            in.ProductID,
		TenantID:             in.TenantID,
		AgeRestriction:       in.AgeRestriction,
		IsHazardous:          in.IsHazardous,
		HazardClass:          in.HazardClass,
		IsPharmacy:           in.IsPharmacy,
		RequiresPrescription: in.RequiresPrescription,
		IsFood:               in.IsFood,
		IsOrganic:            in.IsOrganic,
		IsHalal:              in.IsHalal,
		IsVegan:              in.IsVegan,
		IsGlutenFree:         in.IsGlutenFree,
		RestrictedCountries:  in.RestrictedCountries,
		AllowedCountries:     in.AllowedCountries,
		Certificates:         in.Certificates,
		Metadata:             map[string]any{},
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	if existing, err := d.Compliance.GetByProduct(ctx, in.TenantID, in.ProductID); err == nil {
		c.ID = existing.ID
		c.CreatedAt = existing.CreatedAt
	}
	if err := c.Validate(); err != nil {
		return domain.ProductCompliance{}, err
	}
	if err := d.Compliance.Upsert(ctx, c); err != nil {
		return domain.ProductCompliance{}, err
	}
	return c, nil
}

// GetCompliance returns compliance for a product.
func (d *Deps) GetCompliance(ctx context.Context, tenantID, productID uuid.UUID) (domain.ProductCompliance, error) {
	return d.Compliance.GetByProduct(ctx, tenantID, productID)
}
