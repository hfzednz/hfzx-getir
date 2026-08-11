CREATE TABLE offline_manifests (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     UUID NOT NULL,
    region        TEXT NOT NULL,
    version       TEXT NOT NULL,
    url           TEXT NOT NULL,
    size_bytes    BIGINT NOT NULL DEFAULT 0,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_offline_manifests_tenant_region UNIQUE (tenant_id, region),
    CONSTRAINT chk_offline_size CHECK (size_bytes >= 0),
    CONSTRAINT chk_offline_region CHECK (region <> '')
);

COMMENT ON TABLE offline_manifests IS 'Offline map package manifests (metadata only; tiles not served).';
