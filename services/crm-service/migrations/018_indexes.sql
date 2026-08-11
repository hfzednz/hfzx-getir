-- indexes
CREATE INDEX IF NOT EXISTS idx_tickets_tenant_status ON tickets (tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_tickets_tenant_customer ON tickets (tenant_id, customer_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_tickets_tenant_idempotency
    ON tickets (tenant_id, idempotency_key) WHERE idempotency_key IS NOT NULL AND idempotency_key <> '';
CREATE INDEX IF NOT EXISTS idx_ticket_events_ticket ON ticket_events (tenant_id, ticket_id, created_at);
CREATE INDEX IF NOT EXISTS idx_ticket_notes_ticket ON ticket_notes (tenant_id, ticket_id);
CREATE INDEX IF NOT EXISTS idx_conversations_tenant_customer ON conversations (tenant_id, customer_id);
CREATE INDEX IF NOT EXISTS idx_messages_conversation ON messages (tenant_id, conversation_id, created_at);
CREATE INDEX IF NOT EXISTS idx_agents_tenant ON agents (tenant_id);
CREATE INDEX IF NOT EXISTS idx_kb_articles_tenant_status ON kb_articles (tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_cases_tenant_customer ON cases (tenant_id, customer_id);
CREATE INDEX IF NOT EXISTS idx_csat_tenant_customer ON csat_responses (tenant_id, customer_id);
CREATE INDEX IF NOT EXISTS idx_feedback_tenant ON feedback (tenant_id, kind);
CREATE INDEX IF NOT EXISTS idx_escalations_ticket ON escalations (tenant_id, ticket_id);
CREATE INDEX IF NOT EXISTS idx_outbox_pending ON outbox (status, created_at) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_attachments_owner ON attachments_meta (tenant_id, owner_type, owner_id);
