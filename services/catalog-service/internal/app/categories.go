package app

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/catalog-service/internal/domain"
)

// CreateCategoryInput creates a taxonomy node.
type CreateCategoryInput struct {
	TenantID    uuid.UUID
	ParentID    *uuid.UUID
	Name        string
	Slug        string
	Kind        domain.CategoryKind
	Description string
	SortOrder   int
}

// CreateCategory inserts a category with materialized path.
func (d *Deps) CreateCategory(ctx context.Context, in CreateCategoryInput) (domain.Category, error) {
	if in.Kind == "" {
		in.Kind = domain.CategoryKindStandard
	}
	now := d.now()
	id := d.newID()
	depth := 0
	path := domain.BuildCategoryPath("", id)
	var ancestors []uuid.UUID

	if in.ParentID != nil {
		parent, err := d.Categories.GetByID(ctx, in.TenantID, *in.ParentID)
		if err != nil {
			return domain.Category{}, err
		}
		depth = parent.Depth + 1
		path = domain.BuildCategoryPath(parent.Path, id)
		ids, err := domain.ParseCategoryPathIDs(parent.Path)
		if err != nil {
			return domain.Category{}, err
		}
		ancestors = append(ancestors, ids...)
		ancestors = append(ancestors, parent.ID)
	}

	c := domain.Category{
		ID:          id,
		TenantID:    in.TenantID,
		ParentID:    in.ParentID,
		Name:        strings.TrimSpace(in.Name),
		Slug:        strings.TrimSpace(in.Slug),
		Kind:        in.Kind,
		Path:        path,
		Depth:       depth,
		SortOrder:   in.SortOrder,
		Description: strings.TrimSpace(in.Description),
		Metadata:    map[string]any{},
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := c.Validate(); err != nil {
		return domain.Category{}, err
	}
	if in.ParentID != nil && domain.WouldCreateCategoryCycle(c.ID, *in.ParentID, ancestors) {
		return domain.Category{}, domain.ErrCategoryCycle
	}
	if _, err := d.Categories.GetBySlug(ctx, in.TenantID, c.Slug); err == nil {
		return domain.Category{}, domain.ErrAlreadyExists
	}
	if err := d.Categories.Create(ctx, c); err != nil {
		return domain.Category{}, err
	}
	d.publishEvent(ctx, domain.EventCategoryChanged, in.TenantID, uuid.Nil, map[string]any{"categoryId": c.ID, "action": "created"})
	return c, nil
}

// GetCategory returns a category by id.
func (d *Deps) GetCategory(ctx context.Context, tenantID, categoryID uuid.UUID) (domain.Category, error) {
	return d.Categories.GetByID(ctx, tenantID, categoryID)
}

// ListCategories returns the tenant category tree flat list.
func (d *Deps) ListCategories(ctx context.Context, tenantID uuid.UUID) ([]domain.Category, error) {
	return d.Categories.ListByTenant(ctx, tenantID)
}

// AssignProductCategory links a product to a category.
func (d *Deps) AssignProductCategory(ctx context.Context, tenantID, productID, categoryID uuid.UUID, isPrimary bool, sortOrder int) error {
	if _, err := d.getProduct(ctx, tenantID, productID); err != nil {
		return err
	}
	if _, err := d.Categories.GetByID(ctx, tenantID, categoryID); err != nil {
		return err
	}
	pc := domain.ProductCategory{
		ProductID:  productID,
		CategoryID: categoryID,
		IsPrimary:  isPrimary,
		SortOrder:  sortOrder,
		AssignedAt: d.now(),
	}
	if err := pc.Validate(); err != nil {
		return err
	}
	if err := d.Categories.AssignProduct(ctx, pc); err != nil {
		return err
	}
	d.indexProduct(ctx, tenantID, productID)
	return nil
}

// ListProductCategories returns category memberships for a product.
func (d *Deps) ListProductCategories(ctx context.Context, productID uuid.UUID) ([]domain.ProductCategory, error) {
	return d.Categories.ListProductCategories(ctx, productID)
}
