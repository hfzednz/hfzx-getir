package app

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/catalog-service/internal/domain"
)

// UpsertSEOInput upserts SEO metadata.
type UpsertSEOInput struct {
	TenantID        uuid.UUID
	EntityType      domain.SEOEntityType
	EntityID        uuid.UUID
	Lang            string
	Slug            string
	MetaTitle       string
	MetaDescription string
	CanonicalURL    string
	JSONLD          map[string]any
}

// UpsertSEO upserts SEO for a catalog entity.
func (d *Deps) UpsertSEO(ctx context.Context, in UpsertSEOInput) (domain.SEO, error) {
	if in.JSONLD == nil {
		in.JSONLD = map[string]any{}
	}
	now := d.now()
	existing, err := d.SEO.Get(ctx, in.TenantID, in.EntityType, in.EntityID, in.Lang)
	if err != nil && err != domain.ErrNotFound {
		return domain.SEO{}, err
	}
	s := domain.SEO{
		ID:              d.newID(),
		TenantID:        in.TenantID,
		EntityType:      in.EntityType,
		EntityID:        in.EntityID,
		Lang:            strings.TrimSpace(in.Lang),
		Slug:            strings.TrimSpace(in.Slug),
		MetaTitle:       strings.TrimSpace(in.MetaTitle),
		MetaDescription: strings.TrimSpace(in.MetaDescription),
		CanonicalURL:    strings.TrimSpace(in.CanonicalURL),
		JSONLD:          in.JSONLD,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err == nil {
		s.ID = existing.ID
		s.CreatedAt = existing.CreatedAt
	}
	if err := s.Validate(); err != nil {
		return domain.SEO{}, err
	}
	if err := d.SEO.Upsert(ctx, s); err != nil {
		return domain.SEO{}, err
	}
	return s, nil
}

// GetSEO returns SEO metadata.
func (d *Deps) GetSEO(ctx context.Context, tenantID uuid.UUID, entityType domain.SEOEntityType, entityID uuid.UUID, lang string) (domain.SEO, error) {
	return d.SEO.Get(ctx, tenantID, entityType, entityID, lang)
}
