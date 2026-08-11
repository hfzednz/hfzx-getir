package httpadapter

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/catalog-service/internal/app"
	"github.com/nexora/catalog-service/internal/app/ports"
	"github.com/nexora/catalog-service/internal/domain"
	"github.com/nexora/catalog-service/internal/ratelimit"
)

// Handler serves catalog REST endpoints.
type Handler struct {
	Deps  *app.Deps
	Ready func(*http.Request) error
	Live  func(*http.Request) error
}

// ServerConfig configures the HTTP server.
type ServerConfig struct {
	Addr               string
	Deps               *app.Deps
	Limiter            ratelimit.Limiter
	RateLimitPerMinute int
	CORSOrigins        []string
	Log                *slog.Logger
	Ready              func(*http.Request) error
	Live               func(*http.Request) error
}

// NewHandler returns a fully wired http.Handler.
func NewHandler(cfg ServerConfig) http.Handler {
	h := &Handler{Deps: cfg.Deps, Ready: cfg.Ready, Live: cfg.Live}
	mux := http.NewServeMux()
	const base = "/v1/catalog"

	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /ready", h.ready)
	mux.HandleFunc("GET "+base+"/health", h.health)
	mux.HandleFunc("GET "+base+"/ready", h.ready)

	// Products
	mux.HandleFunc("GET "+base+"/products", tenant(h.listProducts))
	mux.HandleFunc("POST "+base+"/products", tenant(h.createProduct))
	mux.HandleFunc("GET "+base+"/products/{id}", tenant(h.getProduct))
	mux.HandleFunc("PATCH "+base+"/products/{id}", tenant(h.updateProduct))
	mux.HandleFunc("DELETE "+base+"/products/{id}", tenant(h.deleteProduct))
	mux.HandleFunc("POST "+base+"/products/{id}/archive", tenant(h.archiveProduct))

	// Variants & SKUs
	mux.HandleFunc("GET "+base+"/products/{id}/variants", tenant(h.listVariants))
	mux.HandleFunc("POST "+base+"/products/{id}/variants", tenant(h.createVariant))
	mux.HandleFunc("POST "+base+"/variants/{variantId}/skus", tenant(h.addSKU))
	mux.HandleFunc("GET "+base+"/variants/{variantId}/skus", tenant(h.listSKUs))
	mux.HandleFunc("GET "+base+"/skus/lookup", tenant(h.lookupBarcode))

	// Categories
	mux.HandleFunc("GET "+base+"/categories", tenant(h.listCategories))
	mux.HandleFunc("POST "+base+"/categories", tenant(h.createCategory))
	mux.HandleFunc("GET "+base+"/categories/{id}", tenant(h.getCategory))
	mux.HandleFunc("POST "+base+"/products/{id}/categories", tenant(h.assignCategory))

	// Brands
	mux.HandleFunc("GET "+base+"/brands", tenant(h.listBrands))
	mux.HandleFunc("POST "+base+"/brands", tenant(h.createBrand))
	mux.HandleFunc("GET "+base+"/brands/{id}", tenant(h.getBrand))

	// Suppliers
	mux.HandleFunc("GET "+base+"/suppliers", tenant(h.listSuppliers))
	mux.HandleFunc("POST "+base+"/suppliers", tenant(h.createSupplier))
	mux.HandleFunc("GET "+base+"/suppliers/{id}", tenant(h.getSupplier))
	mux.HandleFunc("POST "+base+"/suppliers/{id}/products", tenant(h.linkSupplierProduct))
	mux.HandleFunc("GET "+base+"/suppliers/{id}/products", tenant(h.listSupplierProducts))

	// Attributes
	mux.HandleFunc("GET "+base+"/attributes", tenant(h.listAttributeDefs))
	mux.HandleFunc("POST "+base+"/attributes", tenant(h.createAttributeDef))
	mux.HandleFunc("GET "+base+"/products/{id}/attributes", tenant(h.listProductAttributes))
	mux.HandleFunc("PUT "+base+"/products/{id}/attributes", tenant(h.setProductAttribute))

	// Locales & SEO
	mux.HandleFunc("PUT "+base+"/products/{id}/locales/{lang}", tenant(h.upsertLocale))
	mux.HandleFunc("GET "+base+"/products/{id}/locales", tenant(h.listLocales))
	mux.HandleFunc("PUT "+base+"/seo/{entityType}/{entityId}", tenant(h.upsertSEO))

	// Media
	mux.HandleFunc("GET "+base+"/products/{id}/media", tenant(h.listMedia))
	mux.HandleFunc("POST "+base+"/products/{id}/media", tenant(h.attachMedia))
	mux.HandleFunc("DELETE "+base+"/media/{mediaId}", tenant(h.detachMedia))

	// Bundles & relations
	mux.HandleFunc("PUT "+base+"/products/{id}/bundle", tenant(h.upsertBundle))
	mux.HandleFunc("GET "+base+"/products/{id}/bundle", tenant(h.getBundle))
	mux.HandleFunc("GET "+base+"/products/{id}/relations", tenant(h.listRelations))
	mux.HandleFunc("PUT "+base+"/products/{id}/relations", tenant(h.setRelation))

	// Workflow
	mux.HandleFunc("POST "+base+"/products/{id}/workflow/submit", tenant(h.submitProduct))
	mux.HandleFunc("POST "+base+"/products/{id}/workflow/approve", tenant(h.approveProduct))
	mux.HandleFunc("POST "+base+"/products/{id}/workflow/reject", tenant(h.rejectProduct))
	mux.HandleFunc("POST "+base+"/products/{id}/workflow/publish", tenant(h.publishProduct))
	mux.HandleFunc("POST "+base+"/products/{id}/workflow/hide", tenant(h.hideProduct))
	mux.HandleFunc("GET "+base+"/products/{id}/workflow", tenant(h.listWorkflow))

	// Versions
	mux.HandleFunc("GET "+base+"/products/{id}/versions", tenant(h.listVersions))
	mux.HandleFunc("GET "+base+"/versions/{versionId}", tenant(h.getVersion))
	mux.HandleFunc("GET "+base+"/versions/diff", tenant(h.diffVersions))
	mux.HandleFunc("POST "+base+"/products/{id}/versions/rollback", tenant(h.rollbackVersion))

	// Import
	mux.HandleFunc("POST "+base+"/import/validate", tenant(h.validateImport))
	mux.HandleFunc("GET "+base+"/import/jobs", tenant(h.listImportJobs))
	mux.HandleFunc("GET "+base+"/import/jobs/{jobId}", tenant(h.getImportJob))

	// Search
	mux.HandleFunc("GET "+base+"/search", tenant(h.search))
	mux.HandleFunc("GET "+base+"/search/suggest", tenant(h.suggest))
	mux.HandleFunc("POST "+base+"/search/reindex", tenant(h.reindex))

	// Admin
	mux.HandleFunc("GET "+base+"/admin/explorer", tenant(h.explorer))
	mux.HandleFunc("POST "+base+"/admin/bulk/status", tenant(h.bulkStatus))
	mux.HandleFunc("GET "+base+"/admin/duplicates", tenant(h.duplicates))

	// AI ports
	mux.HandleFunc("POST "+base+"/ai/products/{id}/describe", tenant(h.aiDescribe))
	mux.HandleFunc("POST "+base+"/ai/products/{id}/translate", tenant(h.aiTranslate))
	mux.HandleFunc("POST "+base+"/ai/products/{id}/categorize", tenant(h.aiCategorize))
	mux.HandleFunc("POST "+base+"/ai/products/{id}/quality", tenant(h.aiQuality))

	// Compliance
	mux.HandleFunc("PUT "+base+"/products/{id}/compliance", tenant(h.upsertCompliance))
	mux.HandleFunc("GET "+base+"/products/{id}/compliance", tenant(h.getCompliance))

	return chain(mux,
		requestIDMiddleware,
		recoverMiddleware(cfg.Log),
		loggingMiddleware(cfg.Log),
		corsMiddleware(cfg.CORSOrigins),
		rateLimitMiddleware(cfg.Limiter, cfg.RateLimitPerMinute),
	)
}

// NewServer builds an *http.Server with sensible timeouts.
func NewServer(cfg ServerConfig) *http.Server {
	return &http.Server{
		Addr:              cfg.Addr,
		Handler:           NewHandler(cfg),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

func tenant(next func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantMiddleware(http.HandlerFunc(next)).ServeHTTP(w, r)
	}
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	if h.Live != nil {
		if err := h.Live(r); err != nil {
			writeErr(w, r, err)
			return
		}
	}
	writeOK(w, map[string]string{"status": "ok"})
}

func (h *Handler) ready(w http.ResponseWriter, r *http.Request) {
	if h.Ready != nil {
		if err := h.Ready(r); err != nil {
			writeErr(w, r, err)
			return
		}
	}
	writeOK(w, map[string]string{"status": "ready"})
}

func (h *Handler) tenant(r *http.Request) uuid.UUID {
	tid, _ := TenantIDFromContext(r.Context())
	return tid
}

func parseUUIDParam(r *http.Request, name string) (uuid.UUID, error) {
	return uuid.Parse(r.PathValue(name))
}

func parseActor(r *http.Request) uuid.UUID {
	if v := r.Header.Get("X-Actor-Id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			return id
		}
	}
	return uuid.MustParse("00000000-0000-0000-0000-000000000001")
}

// --- Products ---

func (h *Handler) createProduct(w http.ResponseWriter, r *http.Request) {
	var body struct {
		BrandID     *uuid.UUID         `json:"brandId"`
		Kind        domain.ProductKind `json:"kind"`
		Slug        string             `json:"slug"`
		SKUCode     string             `json:"skuCode"`
		ExternalRef string             `json:"externalRef"`
		Metadata    map[string]any     `json:"metadata"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	p, err := h.Deps.CreateProduct(r.Context(), app.CreateProductInput{
		TenantID: h.tenant(r), BrandID: body.BrandID, Kind: body.Kind,
		Slug: body.Slug, SKUCode: body.SKUCode, ExternalRef: body.ExternalRef, Metadata: body.Metadata,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, map[string]any{"product": p})
}

func (h *Handler) getProduct(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	p, err := h.Deps.GetProduct(r.Context(), h.tenant(r), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"product": p})
}

func (h *Handler) listProducts(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	var status *domain.ProductStatus
	if s := r.URL.Query().Get("status"); s != "" {
		st := domain.ProductStatus(s)
		status = &st
	}
	items, total, err := h.Deps.ListProducts(r.Context(), ports.ProductFilter{
		TenantID: h.tenant(r), Query: r.URL.Query().Get("q"), Status: status, Limit: limit, Offset: offset,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items, "total": total})
}

func (h *Handler) updateProduct(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		BrandID         *uuid.UUID     `json:"brandId"`
		Slug            *string        `json:"slug"`
		SKUCode         *string        `json:"skuCode"`
		ExternalRef     *string        `json:"externalRef"`
		GTINBase        *string        `json:"gtinBase"`
		ManufacturerSKU *string        `json:"manufacturerSku"`
		Metadata        map[string]any `json:"metadata"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	p, err := h.Deps.UpdateProduct(r.Context(), app.UpdateProductInput{
		TenantID: h.tenant(r), ProductID: id, BrandID: body.BrandID, Slug: body.Slug,
		SKUCode: body.SKUCode, ExternalRef: body.ExternalRef, GTINBase: body.GTINBase,
		ManufacturerSKU: body.ManufacturerSKU, Metadata: body.Metadata,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"product": p})
}

func (h *Handler) deleteProduct(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	if err := h.Deps.DeleteProduct(r.Context(), h.tenant(r), id); err != nil {
		writeErr(w, r, err)
		return
	}
	writeNoContent(w)
}

func (h *Handler) archiveProduct(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	p, err := h.Deps.ArchiveProduct(r.Context(), h.tenant(r), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"product": p})
}

// --- Variants / SKUs ---

func (h *Handler) createVariant(w http.ResponseWriter, r *http.Request) {
	productID, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		SKUCode      string         `json:"skuCode"`
		Name         string         `json:"name"`
		OptionValues map[string]any `json:"optionValues"`
		SortOrder    int            `json:"sortOrder"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	v, err := h.Deps.CreateVariant(r.Context(), app.CreateVariantInput{
		TenantID: h.tenant(r), ProductID: productID, SKUCode: body.SKUCode,
		Name: body.Name, OptionValues: body.OptionValues, SortOrder: body.SortOrder,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, map[string]any{"variant": v})
}

func (h *Handler) listVariants(w http.ResponseWriter, r *http.Request) {
	productID, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	items, err := h.Deps.ListVariants(r.Context(), h.tenant(r), productID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) addSKU(w http.ResponseWriter, r *http.Request) {
	variantID, err := parseUUIDParam(r, "variantId")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		Type      domain.SKUIdentifierType `json:"type"`
		Value     string                   `json:"value"`
		IsPrimary bool                     `json:"isPrimary"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	s, err := h.Deps.AddSKUIdentifier(r.Context(), app.AddSKUIdentifierInput{
		TenantID: h.tenant(r), VariantID: variantID, Type: body.Type, Value: body.Value, IsPrimary: body.IsPrimary,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, map[string]any{"sku": s})
}

func (h *Handler) listSKUs(w http.ResponseWriter, r *http.Request) {
	variantID, err := parseUUIDParam(r, "variantId")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	items, err := h.Deps.ListSKUIdentifiers(r.Context(), h.tenant(r), variantID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) lookupBarcode(w http.ResponseWriter, r *http.Request) {
	typ := domain.SKUIdentifierType(r.URL.Query().Get("type"))
	value := r.URL.Query().Get("value")
	if typ == "" || value == "" {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	s, v, p, err := h.Deps.FindByBarcode(r.Context(), h.tenant(r), typ, value)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"sku": s, "variant": v, "product": p})
}

// --- Categories ---

func (h *Handler) createCategory(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ParentID    *uuid.UUID         `json:"parentId"`
		Name        string             `json:"name"`
		Slug        string             `json:"slug"`
		Kind        domain.CategoryKind `json:"kind"`
		Description string             `json:"description"`
		SortOrder   int                `json:"sortOrder"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	c, err := h.Deps.CreateCategory(r.Context(), app.CreateCategoryInput{
		TenantID: h.tenant(r), ParentID: body.ParentID, Name: body.Name, Slug: body.Slug,
		Kind: body.Kind, Description: body.Description, SortOrder: body.SortOrder,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, map[string]any{"category": c})
}

func (h *Handler) getCategory(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	c, err := h.Deps.GetCategory(r.Context(), h.tenant(r), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"category": c})
}

func (h *Handler) listCategories(w http.ResponseWriter, r *http.Request) {
	items, err := h.Deps.ListCategories(r.Context(), h.tenant(r))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) assignCategory(w http.ResponseWriter, r *http.Request) {
	productID, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		CategoryID uuid.UUID `json:"categoryId"`
		IsPrimary  bool      `json:"isPrimary"`
		SortOrder  int       `json:"sortOrder"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	if err := h.Deps.AssignProductCategory(r.Context(), h.tenant(r), productID, body.CategoryID, body.IsPrimary, body.SortOrder); err != nil {
		writeErr(w, r, err)
		return
	}
	writeNoContent(w)
}

// --- Brands ---

func (h *Handler) createBrand(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
		LogoURL     string `json:"logoUrl"`
		WebsiteURL  string `json:"websiteUrl"`
		CountryCode string `json:"countryCode"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	b, err := h.Deps.CreateBrand(r.Context(), app.CreateBrandInput{
		TenantID: h.tenant(r), Name: body.Name, Slug: body.Slug, Description: body.Description,
		LogoURL: body.LogoURL, WebsiteURL: body.WebsiteURL, CountryCode: body.CountryCode,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, map[string]any{"brand": b})
}

func (h *Handler) getBrand(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	b, err := h.Deps.GetBrand(r.Context(), h.tenant(r), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"brand": b})
}

func (h *Handler) listBrands(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	items, total, err := h.Deps.ListBrands(r.Context(), h.tenant(r), limit, offset)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items, "total": total})
}

// --- Suppliers ---

func (h *Handler) createSupplier(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Code         string `json:"code"`
		Name         string `json:"name"`
		ContactEmail string `json:"contactEmail"`
		ContactPhone string `json:"contactPhone"`
		CountryCode  string `json:"countryCode"`
		ExternalRef  string `json:"externalRef"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	s, err := h.Deps.CreateSupplier(r.Context(), app.CreateSupplierInput{
		TenantID: h.tenant(r), Code: body.Code, Name: body.Name,
		ContactEmail: body.ContactEmail, ContactPhone: body.ContactPhone,
		CountryCode: body.CountryCode, ExternalRef: body.ExternalRef,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, map[string]any{"supplier": s})
}

func (h *Handler) getSupplier(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	s, err := h.Deps.GetSupplier(r.Context(), h.tenant(r), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"supplier": s})
}

func (h *Handler) listSuppliers(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	items, total, err := h.Deps.ListSuppliers(r.Context(), h.tenant(r), limit, offset)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items, "total": total})
}

func (h *Handler) linkSupplierProduct(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		ProductID        uuid.UUID  `json:"productId"`
		VariantID        *uuid.UUID `json:"variantId"`
		SupplierSKU      string     `json:"supplierSku"`
		CostHintMinor    *int64     `json:"costHintMinor"`
		CostHintCurrency string     `json:"costHintCurrency"`
		LeadTimeDays     *int       `json:"leadTimeDays"`
		MOQ              *int       `json:"moq"`
		IsPreferred      bool       `json:"isPreferred"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	sp, err := h.Deps.LinkSupplierProduct(r.Context(), app.LinkSupplierProductInput{
		TenantID: h.tenant(r), SupplierID: id, ProductID: body.ProductID, VariantID: body.VariantID,
		SupplierSKU: body.SupplierSKU, CostHintMinor: body.CostHintMinor, CostHintCurrency: body.CostHintCurrency,
		LeadTimeDays: body.LeadTimeDays, MOQ: body.MOQ, IsPreferred: body.IsPreferred,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, map[string]any{"link": sp})
}

func (h *Handler) listSupplierProducts(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	items, err := h.Deps.ListSupplierProducts(r.Context(), h.tenant(r), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

// --- Attributes ---

func (h *Handler) createAttributeDef(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Code          string               `json:"code"`
		Name          string               `json:"name"`
		Description   string               `json:"description"`
		Type          domain.AttributeType `json:"type"`
		Schema        map[string]any       `json:"schema"`
		IsRequired    bool                 `json:"isRequired"`
		IsFilterable  bool                 `json:"isFilterable"`
		IsVariantAxis bool                 `json:"isVariantAxis"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	def, err := h.Deps.CreateAttributeDef(r.Context(), app.CreateAttributeDefInput{
		TenantID: h.tenant(r), Code: body.Code, Name: body.Name, Description: body.Description,
		Type: body.Type, Schema: body.Schema, IsRequired: body.IsRequired,
		IsFilterable: body.IsFilterable, IsVariantAxis: body.IsVariantAxis,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, map[string]any{"attributeDef": def})
}

func (h *Handler) listAttributeDefs(w http.ResponseWriter, r *http.Request) {
	items, err := h.Deps.ListAttributeDefs(r.Context(), h.tenant(r))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) setProductAttribute(w http.ResponseWriter, r *http.Request) {
	productID, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		AttributeDefID uuid.UUID      `json:"attributeDefId"`
		Value          map[string]any `json:"value"`
		Locale         string         `json:"locale"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	a, err := h.Deps.SetProductAttribute(r.Context(), app.SetProductAttributeInput{
		TenantID: h.tenant(r), ProductID: productID, AttributeDefID: body.AttributeDefID,
		Value: body.Value, Locale: body.Locale,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"attribute": a})
}

func (h *Handler) listProductAttributes(w http.ResponseWriter, r *http.Request) {
	productID, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	items, err := h.Deps.ListProductAttributes(r.Context(), h.tenant(r), productID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

// --- Locales / SEO ---

func (h *Handler) upsertLocale(w http.ResponseWriter, r *http.Request) {
	productID, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	lang := r.PathValue("lang")
	var body struct {
		Title       string `json:"title"`
		Subtitle    string `json:"subtitle"`
		Description string `json:"description"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	l, err := h.Deps.UpsertProductLocale(r.Context(), app.UpsertProductLocaleInput{
		TenantID: h.tenant(r), ProductID: productID, Lang: lang,
		Title: body.Title, Subtitle: body.Subtitle, Description: body.Description,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"locale": l})
}

func (h *Handler) listLocales(w http.ResponseWriter, r *http.Request) {
	productID, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	items, err := h.Deps.ListProductLocales(r.Context(), h.tenant(r), productID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) upsertSEO(w http.ResponseWriter, r *http.Request) {
	entityType := domain.SEOEntityType(r.PathValue("entityType"))
	entityID, err := parseUUIDParam(r, "entityId")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		Lang            string         `json:"lang"`
		Slug            string         `json:"slug"`
		MetaTitle       string         `json:"metaTitle"`
		MetaDescription string         `json:"metaDescription"`
		CanonicalURL    string         `json:"canonicalUrl"`
		JSONLD          map[string]any `json:"jsonLd"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	s, err := h.Deps.UpsertSEO(r.Context(), app.UpsertSEOInput{
		TenantID: h.tenant(r), EntityType: entityType, EntityID: entityID, Lang: body.Lang,
		Slug: body.Slug, MetaTitle: body.MetaTitle, MetaDescription: body.MetaDescription,
		CanonicalURL: body.CanonicalURL, JSONLD: body.JSONLD,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"seo": s})
}

// --- Media ---

func (h *Handler) attachMedia(w http.ResponseWriter, r *http.Request) {
	productID, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		MediaAssetID uuid.UUID        `json:"mediaAssetId"`
		VariantID    *uuid.UUID       `json:"variantId"`
		Kind         domain.MediaKind `json:"kind"`
		SortOrder    int              `json:"sortOrder"`
		AltText      string           `json:"altText"`
		IsPrimary    bool             `json:"isPrimary"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	m, err := h.Deps.AttachMedia(r.Context(), app.AttachMediaInput{
		TenantID: h.tenant(r), ProductID: productID, VariantID: body.VariantID,
		MediaAssetID: body.MediaAssetID, Kind: body.Kind, SortOrder: body.SortOrder,
		AltText: body.AltText, IsPrimary: body.IsPrimary,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, map[string]any{"media": m})
}

func (h *Handler) listMedia(w http.ResponseWriter, r *http.Request) {
	productID, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	items, err := h.Deps.ListProductMedia(r.Context(), h.tenant(r), productID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) detachMedia(w http.ResponseWriter, r *http.Request) {
	mediaID, err := parseUUIDParam(r, "mediaId")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	if err := h.Deps.DetachMedia(r.Context(), h.tenant(r), mediaID); err != nil {
		writeErr(w, r, err)
		return
	}
	writeNoContent(w)
}

// --- Bundles / Relations ---

func (h *Handler) upsertBundle(w http.ResponseWriter, r *http.Request) {
	productID, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		Composition domain.BundleCompositionType `json:"composition"`
		Name        string                       `json:"name"`
		Items       []app.BundleItemInput        `json:"items"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	b, items, err := h.Deps.UpsertBundle(r.Context(), app.UpsertBundleInput{
		TenantID: h.tenant(r), ProductID: productID, Composition: body.Composition, Name: body.Name, Items: body.Items,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"bundle": b, "items": items})
}

func (h *Handler) getBundle(w http.ResponseWriter, r *http.Request) {
	productID, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	b, items, err := h.Deps.GetBundle(r.Context(), h.tenant(r), productID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"bundle": b, "items": items})
}

func (h *Handler) setRelation(w http.ResponseWriter, r *http.Request) {
	sourceID, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		TargetProductID uuid.UUID           `json:"targetProductId"`
		Type            domain.RelationType `json:"type"`
		SortOrder       int                 `json:"sortOrder"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	rel, err := h.Deps.SetProductRelation(r.Context(), app.SetProductRelationInput{
		TenantID: h.tenant(r), SourceProductID: sourceID, TargetProductID: body.TargetProductID,
		Type: body.Type, SortOrder: body.SortOrder,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"relation": rel})
}

func (h *Handler) listRelations(w http.ResponseWriter, r *http.Request) {
	sourceID, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	items, err := h.Deps.ListProductRelations(r.Context(), h.tenant(r), sourceID, nil)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

// --- Workflow ---

func (h *Handler) submitProduct(w http.ResponseWriter, r *http.Request) {
	h.workflowAction(w, r, "submit")
}
func (h *Handler) approveProduct(w http.ResponseWriter, r *http.Request) {
	h.workflowAction(w, r, "approve")
}
func (h *Handler) rejectProduct(w http.ResponseWriter, r *http.Request) {
	h.workflowAction(w, r, "reject")
}
func (h *Handler) hideProduct(w http.ResponseWriter, r *http.Request) {
	h.workflowAction(w, r, "hide")
}

func (h *Handler) workflowAction(w http.ResponseWriter, r *http.Request, action string) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct{ Comment string `json:"comment"` }
	_ = decodeJSON(r, &body)
	actor := parseActor(r)
	var p domain.Product
	switch action {
	case "submit":
		p, err = h.Deps.SubmitProduct(r.Context(), h.tenant(r), id, actor, body.Comment)
	case "approve":
		p, err = h.Deps.ApproveProduct(r.Context(), h.tenant(r), id, actor, body.Comment)
	case "reject":
		p, err = h.Deps.RejectProduct(r.Context(), h.tenant(r), id, actor, body.Comment)
	case "hide":
		p, err = h.Deps.HideProduct(r.Context(), h.tenant(r), id, actor, body.Comment)
	}
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"product": p})
}

func (h *Handler) publishProduct(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct{ Comment string `json:"comment"` }
	_ = decodeJSON(r, &body)
	p, ver, err := h.Deps.PublishProduct(r.Context(), h.tenant(r), id, parseActor(r), body.Comment)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"product": p, "version": ver})
}

func (h *Handler) listWorkflow(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	items, err := h.Deps.ListWorkflowActions(r.Context(), h.tenant(r), id, 50)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

// --- Versions ---

func (h *Handler) listVersions(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	items, err := h.Deps.ListProductVersions(r.Context(), h.tenant(r), id, 20, 0)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) getVersion(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "versionId")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	v, err := h.Deps.GetProductVersion(r.Context(), h.tenant(r), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"version": v})
}

func (h *Handler) diffVersions(w http.ResponseWriter, r *http.Request) {
	a, err := uuid.Parse(r.URL.Query().Get("a"))
	b, err2 := uuid.Parse(r.URL.Query().Get("b"))
	if err != nil || err2 != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	diff, err := h.Deps.DiffProductVersions(r.Context(), h.tenant(r), a, b)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"diff": diff})
}

func (h *Handler) rollbackVersion(w http.ResponseWriter, r *http.Request) {
	productID, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		VersionID uuid.UUID `json:"versionId"`
		Comment   string    `json:"comment"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	p, action, err := h.Deps.RollbackProduct(r.Context(), app.RollbackProductInput{
		TenantID: h.tenant(r), ProductID: productID, VersionID: body.VersionID,
		ActorID: parseActor(r), Comment: body.Comment,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"product": p, "action": action})
}

// --- Import ---

func (h *Handler) validateImport(w http.ResponseWriter, r *http.Request) {
	job, err := h.Deps.ValidateImportCSV(r.Context(), h.tenant(r), r.Body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"job": job})
}

func (h *Handler) getImportJob(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "jobId")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	job, err := h.Deps.GetImportJob(r.Context(), h.tenant(r), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"job": job})
}

func (h *Handler) listImportJobs(w http.ResponseWriter, r *http.Request) {
	items, err := h.Deps.ListImportJobs(r.Context(), h.tenant(r), 20, 0)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

// --- Search ---

func (h *Handler) search(w http.ResponseWriter, r *http.Request) {
	var status *domain.ProductStatus
	if s := r.URL.Query().Get("status"); s != "" {
		st := domain.ProductStatus(s)
		status = &st
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	res, err := h.Deps.SearchProducts(r.Context(), ports.SearchQuery{
		TenantID: h.tenant(r), Query: r.URL.Query().Get("q"), Status: status, Limit: limit,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, res)
}

func (h *Handler) suggest(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := h.Deps.SuggestProducts(r.Context(), h.tenant(r), r.URL.Query().Get("prefix"), limit)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"suggestions": items})
}

func (h *Handler) reindex(w http.ResponseWriter, r *http.Request) {
	if err := h.Deps.ReindexTenant(r.Context(), h.tenant(r)); err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]string{"status": "accepted"})
}

// --- Admin ---

func (h *Handler) explorer(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, total, err := h.Deps.ExplorerSearch(r.Context(), app.ExplorerFilter{
		TenantID: h.tenant(r), Query: r.URL.Query().Get("q"), Limit: limit,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items, "total": total})
}

func (h *Handler) bulkStatus(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ProductIDs []uuid.UUID         `json:"productIds"`
		Status     domain.ProductStatus `json:"status"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	items, errs := h.Deps.BulkUpdateStatus(r.Context(), app.BulkUpdateStatusInput{
		TenantID: h.tenant(r), ProductIDs: body.ProductIDs, ToStatus: body.Status, ActorID: parseActor(r),
	})
	writeOK(w, map[string]any{"updated": items, "errors": len(errs)})
}

func (h *Handler) duplicates(w http.ResponseWriter, r *http.Request) {
	items, err := h.Deps.FindDuplicates(r.Context(), h.tenant(r))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"groups": items})
}

// --- AI ---

func (h *Handler) aiDescribe(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	res, err := h.Deps.DescribeProduct(r.Context(), h.tenant(r), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, res)
}

func (h *Handler) aiTranslate(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "tr"
	}
	res, err := h.Deps.TranslateProduct(r.Context(), h.tenant(r), id, lang)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, res)
}

func (h *Handler) aiCategorize(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	res, err := h.Deps.CategorizeProduct(r.Context(), h.tenant(r), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, res)
}

func (h *Handler) aiQuality(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	res, err := h.Deps.QualityScoreProduct(r.Context(), h.tenant(r), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, res)
}

// --- Compliance ---

func (h *Handler) upsertCompliance(w http.ResponseWriter, r *http.Request) {
	productID, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body app.UpsertComplianceInput
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = h.tenant(r)
	body.ProductID = productID
	c, err := h.Deps.UpsertCompliance(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"compliance": c})
}

func (h *Handler) getCompliance(w http.ResponseWriter, r *http.Request) {
	productID, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	c, err := h.Deps.GetCompliance(r.Context(), h.tenant(r), productID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"compliance": c})
}
