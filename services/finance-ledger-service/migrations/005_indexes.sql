-- finance-ledger-service: indexes
CREATE INDEX IF NOT EXISTS idx_coa_tenant ON chart_of_accounts (tenant_id);
CREATE INDEX IF NOT EXISTS idx_journals_tenant_created ON journals (tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_journals_tenant_status ON journals (tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_journal_lines_journal ON journal_lines (journal_id);
CREATE INDEX IF NOT EXISTS idx_journal_lines_account ON journal_lines (tenant_id, account_id);
CREATE INDEX IF NOT EXISTS idx_invoices_tenant_created ON invoices (tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_invoice_lines_invoice ON invoice_lines (invoice_id);
CREATE INDEX IF NOT EXISTS idx_credit_notes_invoice ON credit_notes (tenant_id, invoice_id);
CREATE INDEX IF NOT EXISTS idx_ledger_outbox_pending ON ledger_outbox (status, created_at)
    WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_ledger_events_entity ON ledger_events (tenant_id, entity_id, occurred_at);
