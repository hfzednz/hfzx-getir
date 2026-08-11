-- Indexes for price waterfall resolution, tax, dynamic, audit, outbox.
CREATE INDEX idx_price_books_tenant ON price_books (tenant_id);

CREATE INDEX idx_price_entries_tenant_variant
    ON price_entries (tenant_id, variant_id);

CREATE INDEX idx_price_entries_book
    ON price_entries (tenant_id, price_book_id);

CREATE INDEX idx_price_entries_scope
    ON price_entries (tenant_id, scope, scope_id);

CREATE INDEX idx_price_entries_validity
    ON price_entries (tenant_id, variant_id, valid_from, valid_to);

CREATE INDEX idx_tax_rules_tenant_active
    ON tax_rules (tenant_id, active, priority DESC);

CREATE INDEX idx_tax_rules_region
    ON tax_rules (tenant_id, region_id)
    WHERE region_id IS NOT NULL;

CREATE INDEX idx_dynamic_rules_tenant_active
    ON dynamic_rules (tenant_id, active, priority DESC);

CREATE INDEX idx_quote_audit_tenant_created
    ON quote_audit (tenant_id, created_at DESC);

CREATE INDEX idx_quote_audit_quote
    ON quote_audit (tenant_id, quote_id);

CREATE INDEX idx_outbox_pending
    ON outbox (status, created_at)
    WHERE status = 'pending';

CREATE INDEX idx_outbox_tenant
    ON outbox (tenant_id, created_at DESC);
