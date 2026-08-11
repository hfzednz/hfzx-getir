-- Media: avatar versions and storage references (actual blobs in media-service).
CREATE TABLE profile_media (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id      UUID NOT NULL REFERENCES customer_profiles (id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL,
    version         INT NOT NULL DEFAULT 1 CHECK (version > 0),
    storage_key     TEXT NOT NULL,
    cdn_url         TEXT NOT NULL DEFAULT '',
    content_type    TEXT NOT NULL DEFAULT '',
    bytes           BIGINT NOT NULL DEFAULT 0 CHECK (bytes >= 0),
    is_current      BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_profile_media_profile_version UNIQUE (profile_id, version)
);

COMMENT ON TABLE profile_media IS 'Avatar/media versions; blobs owned by media-service.';
COMMENT ON COLUMN profile_media.storage_key IS 'Object storage key in media-service.';
COMMENT ON COLUMN profile_media.cdn_url IS 'Public or signed CDN URL snapshot.';
COMMENT ON COLUMN profile_media.is_current IS 'True for the active avatar version.';
