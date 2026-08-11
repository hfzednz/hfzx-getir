-- settlement-service: indexes
CREATE INDEX IF NOT EXISTS idx_settlement_batches_tenant_created
    ON settlement_batches (tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_settlement_batches_tenant_status
    ON settlement_batches (tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_settlement_lines_batch
    ON settlement_lines (batch_id);
CREATE INDEX IF NOT EXISTS idx_settlement_lines_payee
    ON settlement_lines (tenant_id, payee_type, payee_ref);
CREATE INDEX IF NOT EXISTS idx_payout_instructions_batch
    ON payout_instructions (tenant_id, batch_id);
CREATE INDEX IF NOT EXISTS idx_reconciliations_batch
    ON reconciliations (tenant_id, batch_id);
CREATE INDEX IF NOT EXISTS idx_mismatches_batch
    ON mismatches (tenant_id, batch_id);
CREATE INDEX IF NOT EXISTS idx_settlement_outbox_pending
    ON settlement_outbox (status, created_at)
    WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_settlement_events_batch
    ON settlement_events (tenant_id, batch_id, occurred_at);
