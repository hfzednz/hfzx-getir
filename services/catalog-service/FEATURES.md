# Catalog Service — Feature Matrix

| Area | Status | Notes |
|------|--------|-------|
| Products CRUD | ✅ | Draft lifecycle, slug validation |
| Variants & SKU identifiers | ✅ | EAN/UPC/GTIN check digits |
| Barcode lookup | ✅ | `FindByBarcode` + HTTP lookup |
| Categories (tree) | ✅ | Materialized path, cycle guard |
| Brands | ✅ | Slug uniqueness per tenant |
| Attribute defs & values | ✅ | Schema-driven json values |
| Locales & SEO | ✅ | BCP-47 lang tags |
| Media refs | ✅ | media-service port stub |
| Bundles / kits | ✅ | Static composition qty (BOM, not stock) |
| Product relations | ✅ | related, accessory, AI, etc. |
| Workflow | ✅ | submit, approve, reject, publish, hide |
| Versions | ✅ | Snapshot on publish, diff, rollback |
| CSV import validate | ✅ | Required columns: slug, sku_code, title, lang |
| Search index/query/suggest | ✅ | Memory + OpenSearch stub |
| Kafka events | ✅ | Log/noop stub without brokers |
| Admin explorer / bulk / dupes | ✅ | |
| AI describe/translate/categorize/quality | ✅ | Honest stubs |
| Compliance flags | ✅ | Age, pharmacy, food, countries |
| PostgreSQL adapters | 🔶 | Migrations ready; repos stub |
| gRPC serve | 🔶 | Proto defined; listener stub |
| Inventory / pricing / orders | ❌ | Out of scope by design |

## Explicit non-goals

- On-hand or reserved stock (`inventory-service`)
- Sellable prices or tax (`pricing-service`)
- Cart / order lines

## Kafka topics

| Topic | Events |
|-------|--------|
| `catalog.product.lifecycle` | ProductCreated, Updated, Published, Archived |
| `catalog.variant.events` | VariantCreated, Updated |
| `catalog.bundle.events` | BundleCreated, Updated |
| `catalog.media.events` | MediaAttached |
| `catalog.category.events` | CategoryChanged |
| `catalog.brand.events` | BrandChanged |
| `catalog.index.commands` | ReindexProduct |

## Product status machine

`draft → pending_review → approved → published`  
Side states: `hidden`, `archived`, `deleted`, `scheduled`

Invalid example: `draft → published` (returns `invalid_transition`).
