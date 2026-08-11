-- Wallet indexes.
CREATE INDEX idx_wallets_tenant_principal
    ON wallets (tenant_id, principal_id);

CREATE INDEX idx_wallet_accounts_wallet
    ON wallet_accounts (wallet_id);

CREATE INDEX idx_wallet_entries_wallet_created
    ON wallet_entries (wallet_id, created_at DESC);

CREATE INDEX idx_wallet_entries_tenant
    ON wallet_entries (tenant_id, created_at DESC);

CREATE INDEX idx_wallet_holds_wallet_status
    ON wallet_holds (wallet_id, status)
    WHERE status = 'active';

CREATE INDEX idx_wallet_transfers_tenant
    ON wallet_transfers (tenant_id, created_at DESC);

CREATE INDEX idx_wallet_limits_wallet
    ON wallet_limits (wallet_id, limit_type);

CREATE INDEX idx_wallet_outbox_pending
    ON wallet_outbox (status, created_at)
    WHERE status = 'pending';

CREATE INDEX idx_wallet_audit_wallet
    ON wallet_audit (wallet_id, created_at DESC);
