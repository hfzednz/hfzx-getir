package memory

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/catalog-service/internal/app/ports"
	"github.com/nexora/catalog-service/internal/domain"
)

type ProductRepo struct{ S *Store }

func (r *ProductRepo) Create(_ context.Context, p domain.Product) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Products[p.ID]; ok {
		return domain.ErrAlreadyExists
	}
	key := tenantKey(p.TenantID, p.Slug)
	if _, ok := r.S.ProductSlug[key]; ok {
		return domain.ErrAlreadyExists
	}
	r.S.Products[p.ID] = p
	r.S.ProductSlug[key] = p.ID
	return nil
}

func (r *ProductRepo) Update(_ context.Context, p domain.Product) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	old, ok := r.S.Products[p.ID]
	if !ok {
		return domain.ErrNotFound
	}
	if old.Slug != p.Slug {
		delete(r.S.ProductSlug, tenantKey(p.TenantID, old.Slug))
		key := tenantKey(p.TenantID, p.Slug)
		if id, exists := r.S.ProductSlug[key]; exists && id != p.ID {
			return domain.ErrAlreadyExists
		}
		r.S.ProductSlug[key] = p.ID
	}
	r.S.Products[p.ID] = p
	return nil
}

func (r *ProductRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (domain.Product, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	p, ok := r.S.Products[id]
	if !ok || p.TenantID != tenantID {
		return domain.Product{}, domain.ErrNotFound
	}
	return p, nil
}

func (r *ProductRepo) GetBySlug(_ context.Context, tenantID uuid.UUID, slug string) (domain.Product, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.ProductSlug[tenantKey(tenantID, slug)]
	if !ok {
		return domain.Product{}, domain.ErrNotFound
	}
	return r.S.Products[id], nil
}

func (r *ProductRepo) List(_ context.Context, f ports.ProductFilter) ([]domain.Product, int, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	all := make([]domain.Product, 0)
	q := strings.ToLower(strings.TrimSpace(f.Query))
	for _, p := range r.S.Products {
		if p.TenantID != f.TenantID {
			continue
		}
		if f.Status != nil && p.Status != *f.Status {
			continue
		}
		if f.BrandID != nil && (p.BrandID == nil || *p.BrandID != *f.BrandID) {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(p.Slug), q) &&
			!strings.Contains(strings.ToLower(p.SKUCode), q) {
			continue
		}
		all = append(all, p)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].CreatedAt.Before(all[j].CreatedAt) })
	total := len(all)
	if f.Offset >= total {
		return nil, total, nil
	}
	end := f.Offset + f.Limit
	if end > total {
		end = total
	}
	return all[f.Offset:end], total, nil
}

func (r *ProductRepo) Delete(_ context.Context, tenantID, id uuid.UUID, at time.Time) error {
	p, err := r.GetByID(context.Background(), tenantID, id)
	if err != nil {
		return err
	}
	p.Status = domain.ProductStatusDeleted
	p.DeletedAt = &at
	p.UpdatedAt = at
	return r.Update(context.Background(), p)
}

type VariantRepo struct{ S *Store }

func (r *VariantRepo) Create(_ context.Context, v domain.Variant) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Variants[v.ID] = v
	return nil
}

func (r *VariantRepo) Update(_ context.Context, v domain.Variant) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Variants[v.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.Variants[v.ID] = v
	return nil
}

func (r *VariantRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (domain.Variant, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	v, ok := r.S.Variants[id]
	if !ok || v.TenantID != tenantID {
		return domain.Variant{}, domain.ErrNotFound
	}
	return v, nil
}

func (r *VariantRepo) ListByProduct(_ context.Context, tenantID, productID uuid.UUID) ([]domain.Variant, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.Variant, 0)
	for _, v := range r.S.Variants {
		if v.TenantID == tenantID && v.ProductID == productID {
			out = append(out, v)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].SortOrder < out[j].SortOrder })
	return out, nil
}

func (r *VariantRepo) Delete(_ context.Context, tenantID, id uuid.UUID, at time.Time) error {
	v, err := r.GetByID(context.Background(), tenantID, id)
	if err != nil {
		return err
	}
	v.Status = domain.VariantStatusDeleted
	v.DeletedAt = &at
	return r.Update(context.Background(), v)
}

type SKURepo struct{ S *Store }

func skuValueKey(tenantID uuid.UUID, typ domain.SKUIdentifierType, value string) string {
	return tenantKey(tenantID, string(typ), value)
}

func (r *SKURepo) Create(_ context.Context, s domain.SKUIdentifier) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	vk := skuValueKey(s.TenantID, s.Type, s.Value)
	if _, ok := r.S.SKUByValue[vk]; ok {
		return domain.ErrAlreadyExists
	}
	r.S.SKUs[s.ID] = s
	r.S.SKUByValue[vk] = s.ID
	return nil
}

func (r *SKURepo) Update(_ context.Context, s domain.SKUIdentifier) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.SKUs[s.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.SKUs[s.ID] = s
	return nil
}

func (r *SKURepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (domain.SKUIdentifier, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	s, ok := r.S.SKUs[id]
	if !ok || s.TenantID != tenantID {
		return domain.SKUIdentifier{}, domain.ErrNotFound
	}
	return s, nil
}

func (r *SKURepo) ListByVariant(_ context.Context, tenantID, variantID uuid.UUID) ([]domain.SKUIdentifier, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.SKUIdentifier, 0)
	for _, s := range r.S.SKUs {
		if s.TenantID == tenantID && s.VariantID == variantID {
			out = append(out, s)
		}
	}
	return out, nil
}

func (r *SKURepo) FindByValue(_ context.Context, tenantID uuid.UUID, typ domain.SKUIdentifierType, value string) (domain.SKUIdentifier, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.SKUByValue[skuValueKey(tenantID, typ, value)]
	if !ok {
		return domain.SKUIdentifier{}, domain.ErrNotFound
	}
	return r.S.SKUs[id], nil
}

func (r *SKURepo) Delete(_ context.Context, tenantID, id uuid.UUID) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	s, ok := r.S.SKUs[id]
	if !ok || s.TenantID != tenantID {
		return domain.ErrNotFound
	}
	delete(r.S.SKUByValue, skuValueKey(s.TenantID, s.Type, s.Value))
	delete(r.S.SKUs, id)
	return nil
}

type CategoryRepo struct{ S *Store }

func (r *CategoryRepo) Create(_ context.Context, c domain.Category) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	key := tenantKey(c.TenantID, c.Slug)
	if _, ok := r.S.CategorySlug[key]; ok {
		return domain.ErrAlreadyExists
	}
	r.S.Categories[c.ID] = c
	r.S.CategorySlug[key] = c.ID
	return nil
}

func (r *CategoryRepo) Update(_ context.Context, c domain.Category) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Categories[c.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.Categories[c.ID] = c
	return nil
}

func (r *CategoryRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (domain.Category, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	c, ok := r.S.Categories[id]
	if !ok || c.TenantID != tenantID {
		return domain.Category{}, domain.ErrNotFound
	}
	return c, nil
}

func (r *CategoryRepo) GetBySlug(_ context.Context, tenantID uuid.UUID, slug string) (domain.Category, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.CategorySlug[tenantKey(tenantID, slug)]
	if !ok {
		return domain.Category{}, domain.ErrNotFound
	}
	return r.S.Categories[id], nil
}

func (r *CategoryRepo) ListByTenant(_ context.Context, tenantID uuid.UUID) ([]domain.Category, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.Category, 0)
	for _, c := range r.S.Categories {
		if c.TenantID == tenantID {
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Depth != out[j].Depth {
			return out[i].Depth < out[j].Depth
		}
		return out[i].SortOrder < out[j].SortOrder
	})
	return out, nil
}

func (r *CategoryRepo) ListChildren(_ context.Context, tenantID, parentID uuid.UUID) ([]domain.Category, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.Category, 0)
	for _, c := range r.S.Categories {
		if c.TenantID == tenantID && c.ParentID != nil && *c.ParentID == parentID {
			out = append(out, c)
		}
	}
	return out, nil
}

func (r *CategoryRepo) AssignProduct(_ context.Context, pc domain.ProductCategory) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	key := pc.ProductID.String() + ":" + pc.CategoryID.String()
	r.S.ProductCategories[key] = pc
	return nil
}

func (r *CategoryRepo) RemoveProduct(_ context.Context, productID, categoryID uuid.UUID) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	delete(r.S.ProductCategories, productID.String()+":"+categoryID.String())
	return nil
}

func (r *CategoryRepo) ListProductCategories(_ context.Context, productID uuid.UUID) ([]domain.ProductCategory, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.ProductCategory, 0)
	for _, pc := range r.S.ProductCategories {
		if pc.ProductID == productID {
			out = append(out, pc)
		}
	}
	return out, nil
}

func (r *CategoryRepo) ListProductsInCategory(_ context.Context, tenantID, categoryID uuid.UUID, limit, offset int) ([]uuid.UUID, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	ids := make([]uuid.UUID, 0)
	for _, pc := range r.S.ProductCategories {
		if pc.CategoryID == categoryID {
			if p, ok := r.S.Products[pc.ProductID]; ok && p.TenantID == tenantID {
				ids = append(ids, pc.ProductID)
			}
		}
	}
	if offset >= len(ids) {
		return nil, nil
	}
	end := offset + limit
	if end > len(ids) {
		end = len(ids)
	}
	return ids[offset:end], nil
}

type BrandRepo struct{ S *Store }

func (r *BrandRepo) Create(_ context.Context, b domain.Brand) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.BrandSlug[tenantKey(b.TenantID, b.Slug)]; ok {
		return domain.ErrAlreadyExists
	}
	r.S.Brands[b.ID] = b
	r.S.BrandSlug[tenantKey(b.TenantID, b.Slug)] = b.ID
	return nil
}

func (r *BrandRepo) Update(_ context.Context, b domain.Brand) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Brands[b.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.Brands[b.ID] = b
	return nil
}

func (r *BrandRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (domain.Brand, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	b, ok := r.S.Brands[id]
	if !ok || b.TenantID != tenantID {
		return domain.Brand{}, domain.ErrNotFound
	}
	return b, nil
}

func (r *BrandRepo) GetBySlug(_ context.Context, tenantID uuid.UUID, slug string) (domain.Brand, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.BrandSlug[tenantKey(tenantID, slug)]
	if !ok {
		return domain.Brand{}, domain.ErrNotFound
	}
	return r.S.Brands[id], nil
}

func (r *BrandRepo) List(_ context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.Brand, int, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	all := make([]domain.Brand, 0)
	for _, b := range r.S.Brands {
		if b.TenantID == tenantID {
			all = append(all, b)
		}
	}
	total := len(all)
	if offset >= total {
		return nil, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return all[offset:end], total, nil
}

type AttributeRepo struct{ S *Store }

func (r *AttributeRepo) CreateDef(_ context.Context, d domain.AttributeDef) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.AttributeDefCode[tenantKey(d.TenantID, d.Code)]; ok {
		return domain.ErrAlreadyExists
	}
	r.S.AttributeDefs[d.ID] = d
	r.S.AttributeDefCode[tenantKey(d.TenantID, d.Code)] = d.ID
	return nil
}

func (r *AttributeRepo) UpdateDef(_ context.Context, d domain.AttributeDef) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.AttributeDefs[d.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.AttributeDefs[d.ID] = d
	return nil
}

func (r *AttributeRepo) GetDefByID(_ context.Context, tenantID, id uuid.UUID) (domain.AttributeDef, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	d, ok := r.S.AttributeDefs[id]
	if !ok || d.TenantID != tenantID {
		return domain.AttributeDef{}, domain.ErrNotFound
	}
	return d, nil
}

func (r *AttributeRepo) GetDefByCode(_ context.Context, tenantID uuid.UUID, code string) (domain.AttributeDef, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.AttributeDefCode[tenantKey(tenantID, code)]
	if !ok {
		return domain.AttributeDef{}, domain.ErrNotFound
	}
	return r.S.AttributeDefs[id], nil
}

func (r *AttributeRepo) ListDefs(_ context.Context, tenantID uuid.UUID) ([]domain.AttributeDef, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.AttributeDef, 0)
	for _, d := range r.S.AttributeDefs {
		if d.TenantID == tenantID {
			out = append(out, d)
		}
	}
	return out, nil
}

func (r *AttributeRepo) UpsertProductAttribute(_ context.Context, a domain.ProductAttribute) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	for id, existing := range r.S.ProductAttributes {
		if existing.ProductID == a.ProductID && existing.AttributeDefID == a.AttributeDefID && existing.Locale == a.Locale {
			a.ID = id
			break
		}
	}
	r.S.ProductAttributes[a.ID] = a
	return nil
}

func (r *AttributeRepo) ListProductAttributes(_ context.Context, tenantID, productID uuid.UUID) ([]domain.ProductAttribute, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.ProductAttribute, 0)
	for _, a := range r.S.ProductAttributes {
		if a.TenantID == tenantID && a.ProductID == productID {
			out = append(out, a)
		}
	}
	return out, nil
}

func (r *AttributeRepo) DeleteProductAttribute(_ context.Context, tenantID, id uuid.UUID) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	a, ok := r.S.ProductAttributes[id]
	if !ok || a.TenantID != tenantID {
		return domain.ErrNotFound
	}
	delete(r.S.ProductAttributes, id)
	return nil
}

type LocaleRepo struct{ S *Store }

func localeKey(tenantID, productID uuid.UUID, lang string) string {
	return tenantKey(tenantID, productID.String(), lang)
}

func (r *LocaleRepo) Upsert(_ context.Context, l domain.ProductLocale) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	key := localeKey(l.TenantID, l.ProductID, l.Lang)
	if id, ok := r.S.LocaleKey[key]; ok {
		l.ID = id
	}
	r.S.Locales[l.ID] = l
	r.S.LocaleKey[key] = l.ID
	return nil
}

func (r *LocaleRepo) Get(_ context.Context, tenantID, productID uuid.UUID, lang string) (domain.ProductLocale, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.LocaleKey[localeKey(tenantID, productID, lang)]
	if !ok {
		return domain.ProductLocale{}, domain.ErrNotFound
	}
	return r.S.Locales[id], nil
}

func (r *LocaleRepo) ListByProduct(_ context.Context, tenantID, productID uuid.UUID) ([]domain.ProductLocale, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.ProductLocale, 0)
	for _, l := range r.S.Locales {
		if l.TenantID == tenantID && l.ProductID == productID {
			out = append(out, l)
		}
	}
	return out, nil
}

func (r *LocaleRepo) Delete(_ context.Context, tenantID, id uuid.UUID) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	l, ok := r.S.Locales[id]
	if !ok || l.TenantID != tenantID {
		return domain.ErrNotFound
	}
	delete(r.S.LocaleKey, localeKey(l.TenantID, l.ProductID, l.Lang))
	delete(r.S.Locales, id)
	return nil
}

type SEORepo struct{ S *Store }

func seoKey(tenantID uuid.UUID, et domain.SEOEntityType, entityID uuid.UUID, lang string) string {
	return tenantKey(tenantID, string(et), entityID.String(), lang)
}

func (r *SEORepo) Upsert(_ context.Context, s domain.SEO) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	key := seoKey(s.TenantID, s.EntityType, s.EntityID, s.Lang)
	if id, ok := r.S.SEOKey[key]; ok {
		s.ID = id
	}
	r.S.SEO[s.ID] = s
	r.S.SEOKey[key] = s.ID
	return nil
}

func (r *SEORepo) Get(_ context.Context, tenantID uuid.UUID, entityType domain.SEOEntityType, entityID uuid.UUID, lang string) (domain.SEO, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.SEOKey[seoKey(tenantID, entityType, entityID, lang)]
	if !ok {
		return domain.SEO{}, domain.ErrNotFound
	}
	return r.S.SEO[id], nil
}

func (r *SEORepo) ListByEntity(_ context.Context, tenantID uuid.UUID, entityType domain.SEOEntityType, entityID uuid.UUID) ([]domain.SEO, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.SEO, 0)
	for _, s := range r.S.SEO {
		if s.TenantID == tenantID && s.EntityType == entityType && s.EntityID == entityID {
			out = append(out, s)
		}
	}
	return out, nil
}

type MediaRepo struct{ S *Store }

func (r *MediaRepo) Create(_ context.Context, m domain.ProductMedia) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Media[m.ID] = m
	return nil
}

func (r *MediaRepo) Update(_ context.Context, m domain.ProductMedia) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Media[m.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.Media[m.ID] = m
	return nil
}

func (r *MediaRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (domain.ProductMedia, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	m, ok := r.S.Media[id]
	if !ok || m.TenantID != tenantID {
		return domain.ProductMedia{}, domain.ErrNotFound
	}
	return m, nil
}

func (r *MediaRepo) ListByProduct(_ context.Context, tenantID, productID uuid.UUID) ([]domain.ProductMedia, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.ProductMedia, 0)
	for _, m := range r.S.Media {
		if m.TenantID == tenantID && m.ProductID == productID {
			out = append(out, m)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].SortOrder < out[j].SortOrder })
	return out, nil
}

func (r *MediaRepo) Delete(_ context.Context, tenantID, id uuid.UUID) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	m, ok := r.S.Media[id]
	if !ok || m.TenantID != tenantID {
		return domain.ErrNotFound
	}
	delete(r.S.Media, id)
	return nil
}

type BundleRepo struct{ S *Store }

func (r *BundleRepo) Upsert(_ context.Context, b domain.Bundle) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Bundles[b.ID] = b
	r.S.BundleByProduct[b.ProductID] = b.ID
	return nil
}

func (r *BundleRepo) GetByProduct(_ context.Context, tenantID, productID uuid.UUID) (domain.Bundle, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.BundleByProduct[productID]
	if !ok {
		return domain.Bundle{}, domain.ErrNotFound
	}
	b := r.S.Bundles[id]
	if b.TenantID != tenantID {
		return domain.Bundle{}, domain.ErrNotFound
	}
	return b, nil
}

func (r *BundleRepo) SetItems(_ context.Context, bundleID uuid.UUID, items []domain.BundleItem) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.BundleItems[bundleID] = items
	return nil
}

func (r *BundleRepo) ListItems(_ context.Context, bundleID uuid.UUID) ([]domain.BundleItem, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	return r.S.BundleItems[bundleID], nil
}

type RelationRepo struct{ S *Store }

func (r *RelationRepo) Upsert(_ context.Context, rel domain.ProductRelation) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	for id, existing := range r.S.Relations {
		if existing.SourceProductID == rel.SourceProductID && existing.TargetProductID == rel.TargetProductID && existing.Type == rel.Type {
			rel.ID = id
			break
		}
	}
	r.S.Relations[rel.ID] = rel
	return nil
}

func (r *RelationRepo) Delete(_ context.Context, tenantID, id uuid.UUID) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	rel, ok := r.S.Relations[id]
	if !ok || rel.TenantID != tenantID {
		return domain.ErrNotFound
	}
	delete(r.S.Relations, id)
	return nil
}

func (r *RelationRepo) ListBySource(_ context.Context, tenantID, sourceID uuid.UUID, typ *domain.RelationType) ([]domain.ProductRelation, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.ProductRelation, 0)
	for _, rel := range r.S.Relations {
		if rel.TenantID == tenantID && rel.SourceProductID == sourceID {
			if typ == nil || rel.Type == *typ {
				out = append(out, rel)
			}
		}
	}
	return out, nil
}

type VersionRepo struct{ S *Store }

func versionCountKey(tenantID, productID uuid.UUID) string {
	return tenantKey(tenantID, productID.String())
}

func (r *VersionRepo) Create(_ context.Context, v domain.ProductVersion) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Versions[v.ID] = v
	key := versionCountKey(v.TenantID, v.ProductID)
	if v.VersionNumber > r.S.VersionCount[key] {
		r.S.VersionCount[key] = v.VersionNumber
	}
	return nil
}

func (r *VersionRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (domain.ProductVersion, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	v, ok := r.S.Versions[id]
	if !ok || v.TenantID != tenantID {
		return domain.ProductVersion{}, domain.ErrNotFound
	}
	return v, nil
}

func (r *VersionRepo) GetLatest(_ context.Context, tenantID, productID uuid.UUID) (domain.ProductVersion, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var latest domain.ProductVersion
	for _, v := range r.S.Versions {
		if v.TenantID == tenantID && v.ProductID == productID {
			if v.VersionNumber > latest.VersionNumber {
				latest = v
			}
		}
	}
	if latest.ID == uuid.Nil {
		return domain.ProductVersion{}, domain.ErrNotFound
	}
	return latest, nil
}

func (r *VersionRepo) ListByProduct(_ context.Context, tenantID, productID uuid.UUID, limit, offset int) ([]domain.ProductVersion, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	all := make([]domain.ProductVersion, 0)
	for _, v := range r.S.Versions {
		if v.TenantID == tenantID && v.ProductID == productID {
			all = append(all, v)
		}
	}
	sort.Slice(all, func(i, j int) bool { return all[i].VersionNumber > all[j].VersionNumber })
	if offset >= len(all) {
		return nil, nil
	}
	end := offset + limit
	if end > len(all) {
		end = len(all)
	}
	return all[offset:end], nil
}

func (r *VersionRepo) NextVersionNumber(_ context.Context, tenantID, productID uuid.UUID) (int, error) {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	key := versionCountKey(tenantID, productID)
	r.S.VersionCount[key]++
	return r.S.VersionCount[key], nil
}

type WorkflowRepo struct{ S *Store }

func (r *WorkflowRepo) CreateAction(_ context.Context, a domain.ApprovalAction) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Workflow = append(r.S.Workflow, a)
	return nil
}

func (r *WorkflowRepo) ListByProduct(_ context.Context, tenantID, productID uuid.UUID, limit int) ([]domain.ApprovalAction, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.ApprovalAction, 0)
	for i := len(r.S.Workflow) - 1; i >= 0; i-- {
		a := r.S.Workflow[i]
		if a.TenantID == tenantID && a.ProductID == productID {
			out = append(out, a)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

type ImportJobRepo struct{ S *Store }

func (r *ImportJobRepo) Create(_ context.Context, j domain.ImportJob) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.ImportJobs[j.ID] = j
	return nil
}

func (r *ImportJobRepo) Update(_ context.Context, j domain.ImportJob) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.ImportJobs[j.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.ImportJobs[j.ID] = j
	return nil
}

func (r *ImportJobRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (domain.ImportJob, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	j, ok := r.S.ImportJobs[id]
	if !ok || j.TenantID != tenantID {
		return domain.ImportJob{}, domain.ErrNotFound
	}
	return j, nil
}

func (r *ImportJobRepo) List(_ context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.ImportJob, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	all := make([]domain.ImportJob, 0)
	for _, j := range r.S.ImportJobs {
		if j.TenantID == tenantID {
			all = append(all, j)
		}
	}
	sort.Slice(all, func(i, j int) bool { return all[i].CreatedAt.After(all[j].CreatedAt) })
	if offset >= len(all) {
		return nil, nil
	}
	end := offset + limit
	if end > len(all) {
		end = len(all)
	}
	return all[offset:end], nil
}

type ComplianceRepo struct{ S *Store }

func (r *ComplianceRepo) Upsert(_ context.Context, c domain.ProductCompliance) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Compliance[c.ID] = c
	r.S.ComplianceByProduct[c.ProductID] = c.ID
	return nil
}

func (r *ComplianceRepo) GetByProduct(_ context.Context, tenantID, productID uuid.UUID) (domain.ProductCompliance, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.ComplianceByProduct[productID]
	if !ok {
		return domain.ProductCompliance{}, domain.ErrNotFound
	}
	c := r.S.Compliance[id]
	if c.TenantID != tenantID {
		return domain.ProductCompliance{}, domain.ErrNotFound
	}
	return c, nil
}

// Ensure repos implement ports at compile time.
var (
	_ ports.ProductRepository     = (*ProductRepo)(nil)
	_ ports.VariantRepository     = (*VariantRepo)(nil)
	_ ports.SKUIdentifierRepository = (*SKURepo)(nil)
	_ ports.CategoryRepository    = (*CategoryRepo)(nil)
	_ ports.BrandRepository       = (*BrandRepo)(nil)
	_ ports.AttributeRepository   = (*AttributeRepo)(nil)
	_ ports.LocaleRepository      = (*LocaleRepo)(nil)
	_ ports.SEORepository         = (*SEORepo)(nil)
	_ ports.MediaRepository       = (*MediaRepo)(nil)
	_ ports.BundleRepository      = (*BundleRepo)(nil)
	_ ports.RelationRepository    = (*RelationRepo)(nil)
	_ ports.VersionRepository     = (*VersionRepo)(nil)
	_ ports.WorkflowRepository    = (*WorkflowRepo)(nil)
	_ ports.ImportJobRepository   = (*ImportJobRepo)(nil)
	_ ports.ComplianceRepository  = (*ComplianceRepo)(nil)
	_ ports.SearchIndexer         = (*SearchIndexer)(nil)
	_ ports.EventPublisher        = (*EventPublisher)(nil)
	_ ports.MediaClient           = (*MediaClient)(nil)
	_ ports.AIClient              = (*AIClient)(nil)
)

// FixedTime helper for tests.
func FixedTime(t time.Time) *Clock { return &Clock{T: t} }
