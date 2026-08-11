-- Performance indexes for catalog read/write and lookup paths.

-- Brands
CREATE INDEX idx_brands_tenant_status ON brands (tenant_id, status)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_brands_tenant_name ON brands (tenant_id, name)
    WHERE deleted_at IS NULL;

-- Categories
CREATE INDEX idx_categories_tenant_parent ON categories (tenant_id, parent_id)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_categories_tenant_kind ON categories (tenant_id, kind)
    WHERE deleted_at IS NULL AND is_active = TRUE;
CREATE INDEX idx_categories_path ON categories (tenant_id, path text_pattern_ops)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_product_categories_category ON product_categories (category_id);
CREATE INDEX idx_product_categories_primary ON product_categories (product_id)
    WHERE is_primary = TRUE;

-- Attribute defs
CREATE INDEX idx_attribute_defs_tenant_type ON attribute_defs (tenant_id, type)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_attribute_defs_filterable ON attribute_defs (tenant_id)
    WHERE is_filterable = TRUE AND deleted_at IS NULL;

-- Products
CREATE INDEX idx_products_tenant_status ON products (tenant_id, status)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_products_tenant_brand ON products (tenant_id, brand_id)
    WHERE deleted_at IS NULL AND brand_id IS NOT NULL;
CREATE INDEX idx_products_tenant_kind ON products (tenant_id, kind)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_products_tenant_external ON products (tenant_id, external_ref)
    WHERE external_ref <> '' AND deleted_at IS NULL;
CREATE INDEX idx_products_scheduled ON products (scheduled_at)
    WHERE status = 'scheduled' AND scheduled_at IS NOT NULL;
CREATE INDEX idx_products_published_at ON products (tenant_id, published_at DESC)
    WHERE status = 'published' AND deleted_at IS NULL;

-- Variants
CREATE INDEX idx_variants_product ON variants (product_id)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_variants_tenant_status ON variants (tenant_id, status)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_variants_tenant_sku ON variants (tenant_id, sku_code)
    WHERE sku_code <> '' AND deleted_at IS NULL;
CREATE INDEX idx_variants_option_values ON variants USING GIN (option_values);

-- SKU identifiers
CREATE INDEX idx_sku_identifiers_variant ON sku_identifiers (variant_id);
CREATE INDEX idx_sku_identifiers_tenant_value ON sku_identifiers (tenant_id, value);
CREATE UNIQUE INDEX uq_sku_identifiers_primary
    ON sku_identifiers (variant_id, type)
    WHERE is_primary = TRUE;

-- Product attributes
CREATE INDEX idx_product_attributes_def ON product_attributes (attribute_def_id);
CREATE INDEX idx_product_attributes_value ON product_attributes USING GIN (value);

-- Locales
CREATE INDEX idx_product_locales_tenant_lang ON product_locales (tenant_id, lang);
CREATE INDEX idx_product_locales_title ON product_locales (tenant_id, lang, title);

-- SEO
CREATE INDEX idx_seo_entity ON seo (entity_type, entity_id);
CREATE INDEX idx_seo_tenant_slug ON seo (tenant_id, slug)
    WHERE slug <> '';

-- Media
CREATE INDEX idx_product_media_product_sort ON product_media (product_id, sort_order);
CREATE INDEX idx_product_media_variant ON product_media (variant_id)
    WHERE variant_id IS NOT NULL;
CREATE INDEX idx_product_media_asset ON product_media (media_asset_id);
CREATE INDEX idx_product_media_primary ON product_media (product_id)
    WHERE is_primary = TRUE;

-- Bundles
CREATE INDEX idx_bundle_items_component ON bundle_items (component_variant_id);
CREATE INDEX idx_bundles_tenant ON bundles (tenant_id);

-- Relations
CREATE INDEX idx_product_relations_source ON product_relations (source_product_id, type, sort_order);
CREATE INDEX idx_product_relations_target ON product_relations (target_product_id, type);
CREATE INDEX idx_product_relations_tenant ON product_relations (tenant_id, type);

-- Suppliers
CREATE INDEX idx_suppliers_tenant_status ON suppliers (tenant_id, status)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_supplier_products_product ON supplier_products (product_id);
CREATE INDEX idx_supplier_products_variant ON supplier_products (variant_id)
    WHERE variant_id IS NOT NULL;
CREATE INDEX idx_supplier_products_sku ON supplier_products (tenant_id, supplier_sku)
    WHERE supplier_sku <> '';

-- Compliance
CREATE INDEX idx_product_compliance_hazard ON product_compliance (tenant_id)
    WHERE is_hazardous = TRUE;
CREATE INDEX idx_product_compliance_pharmacy ON product_compliance (tenant_id)
    WHERE is_pharmacy = TRUE;
CREATE INDEX idx_product_compliance_age ON product_compliance (tenant_id, age_restriction)
    WHERE age_restriction > 0;

-- Versions / workflow
CREATE INDEX idx_product_versions_product ON product_versions (product_id, version_number DESC);
CREATE INDEX idx_approval_actions_product ON approval_actions (product_id, created_at DESC);
CREATE INDEX idx_approval_actions_actor ON approval_actions (actor_id, created_at DESC);
CREATE INDEX idx_approval_actions_tenant_action ON approval_actions (tenant_id, action, created_at DESC);

-- Import jobs
CREATE INDEX idx_import_jobs_tenant_status ON import_jobs (tenant_id, status, created_at DESC);
CREATE INDEX idx_import_jobs_pending ON import_jobs (kind, created_at)
    WHERE status IN ('pending', 'validating', 'running');

-- Collections
CREATE INDEX idx_collections_tenant_kind ON collections (tenant_id, kind)
    WHERE deleted_at IS NULL AND is_active = TRUE;
CREATE INDEX idx_collections_rules ON collections USING GIN (rules)
    WHERE kind = 'smart';
CREATE INDEX idx_collection_products_product ON collection_products (product_id);
CREATE INDEX idx_collection_products_sort ON collection_products (collection_id, sort_order);
