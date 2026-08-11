package app

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/catalog-service/internal/app/ports"
	"github.com/nexora/catalog-service/internal/domain"
)

// Deps aggregates application ports for catalog use cases.
type Deps struct {
	Products    ports.ProductRepository
	Variants    ports.VariantRepository
	SKUs        ports.SKUIdentifierRepository
	Categories  ports.CategoryRepository
	Brands      ports.BrandRepository
	Attributes  ports.AttributeRepository
	Locales     ports.LocaleRepository
	SEO         ports.SEORepository
	Media       ports.MediaRepository
	Bundles     ports.BundleRepository
	Relations   ports.RelationRepository
	Versions    ports.VersionRepository
	Workflow    ports.WorkflowRepository
	ImportJobs  ports.ImportJobRepository
	Compliance  ports.ComplianceRepository
	Suppliers   ports.SupplierRepository
	Search      ports.SearchIndexer
	Events      ports.EventPublisher
	MediaClient ports.MediaClient
	AI          ports.AIClient
	Clock       ports.Clock
	IDs         ports.IDGen
}

func (d *Deps) now() time.Time {
	if d.Clock != nil {
		return d.Clock.Now().UTC()
	}
	return time.Now().UTC()
}

func (d *Deps) newID() uuid.UUID {
	if d.IDs != nil {
		return d.IDs.New()
	}
	return uuid.New()
}

// SystemClock is a real-time Clock.
type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now().UTC() }

// UUIDGen generates random UUIDs.
type UUIDGen struct{}

func (UUIDGen) New() uuid.UUID { return uuid.New() }

func (d *Deps) getProduct(ctx context.Context, tenantID, productID uuid.UUID) (domain.Product, error) {
	p, err := d.Products.GetByID(ctx, tenantID, productID)
	if err != nil {
		return domain.Product{}, err
	}
	if p.Status == domain.ProductStatusDeleted {
		return domain.Product{}, domain.ErrProductDeleted
	}
	return p, nil
}

func (d *Deps) ensureEditable(p domain.Product) error {
	if !p.IsEditable() {
		return domain.ErrProductNotEditable
	}
	return nil
}

func (d *Deps) publishEvent(ctx context.Context, eventType string, tenantID, productID uuid.UUID, payload map[string]any) {
	if d.Events == nil {
		return
	}
	ev := domain.NewDomainEvent(eventType, tenantID, productID, payload)
	topic := domain.TopicForEvent(eventType)
	_ = d.Events.Publish(ctx, topic, productID.String(), ev)
}

func (d *Deps) indexProduct(ctx context.Context, tenantID, productID uuid.UUID) {
	if d.Search == nil {
		return
	}
	p, err := d.Products.GetByID(ctx, tenantID, productID)
	if err != nil {
		return
	}
	doc := d.buildSearchDoc(ctx, p)
	_ = d.Search.IndexProduct(ctx, doc)
	d.publishEvent(ctx, domain.EventReindexProduct, tenantID, productID, map[string]any{"productId": productID})
}

func (d *Deps) buildSearchDoc(ctx context.Context, p domain.Product) ports.SearchDocument {
	doc := ports.SearchDocument{
		ProductID:  p.ID,
		TenantID:   p.TenantID,
		SKU:        p.SKUCode,
		Status:     p.Status,
		Attributes: map[string]any{},
		Locales:    map[string]map[string]string{},
	}
	if p.BrandID != nil {
		if b, err := d.Brands.GetByID(ctx, p.TenantID, *p.BrandID); err == nil {
			doc.Brand = b.Name
		}
	}
	if locs, err := d.Locales.ListByProduct(ctx, p.TenantID, p.ID); err == nil {
		for _, l := range locs {
			doc.Locales[l.Lang] = map[string]string{
				"title":       l.Title,
				"description": l.Description,
			}
			if doc.Title == "" {
				doc.Title = l.Title
			}
		}
	}
	if pcs, err := d.Categories.ListProductCategories(ctx, p.ID); err == nil {
		for _, pc := range pcs {
			doc.CategoryIDs = append(doc.CategoryIDs, pc.CategoryID)
		}
	}
	if attrs, err := d.Attributes.ListProductAttributes(ctx, p.TenantID, p.ID); err == nil {
		for _, a := range attrs {
			doc.Attributes[a.AttributeDefID.String()] = a.Value
		}
	}
	variants, _ := d.Variants.ListByProduct(ctx, p.TenantID, p.ID)
	for _, v := range variants {
		ids, _ := d.SKUs.ListByVariant(ctx, p.TenantID, v.ID)
		for _, id := range ids {
			doc.Barcodes = append(doc.Barcodes, id.Value)
		}
	}
	return doc
}

func (d *Deps) snapshotProduct(ctx context.Context, p domain.Product) map[string]any {
	snap := map[string]any{
		"product": p,
	}
	if locs, err := d.Locales.ListByProduct(ctx, p.TenantID, p.ID); err == nil {
		snap["locales"] = locs
	}
	if vars, err := d.Variants.ListByProduct(ctx, p.TenantID, p.ID); err == nil {
		snap["variants"] = vars
	}
	if attrs, err := d.Attributes.ListProductAttributes(ctx, p.TenantID, p.ID); err == nil {
		snap["attributes"] = attrs
	}
	return snap
}

func diffSnapshots(a, b map[string]any) map[string]any {
	out := map[string]any{}
	rawA, _ := json.Marshal(a)
	rawB, _ := json.Marshal(b)
	if string(rawA) != string(rawB) {
		out["before"] = a
		out["after"] = b
	}
	return out
}
