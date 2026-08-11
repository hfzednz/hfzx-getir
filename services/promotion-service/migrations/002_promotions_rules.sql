-- promotions + rules
CREATE TABLE IF NOT EXISTS promotions (
    id                  UUID PRIMARY KEY,
    tenant_id           UUID NOT NULL,
    campaign_id         UUID NOT NULL REFERENCES campaigns(id),
    name                TEXT NOT NULL,
    type                TEXT NOT NULL CHECK (type IN (
        'percent','fixed','bogo','bxgy','bundle','threshold','free_ship','gift','multibuy'
    )),
    percent_off         INT NOT NULL DEFAULT 0,
    fixed_off_minor     BIGINT NOT NULL DEFAULT 0,
    buy_qty             INT NOT NULL DEFAULT 0,
    get_qty             INT NOT NULL DEFAULT 0,
    threshold_minor     BIGINT NOT NULL DEFAULT 0,
    gift_variant_id     TEXT NOT NULL DEFAULT '',
    max_discount_minor  BIGINT NOT NULL DEFAULT 0,
    priority            INT NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS rules (
    id                      UUID PRIMARY KEY,
    tenant_id               UUID NOT NULL,
    promotion_id            UUID NOT NULL UNIQUE REFERENCES promotions(id),
    priority                INT NOT NULL DEFAULT 0,
    stack_group             TEXT NOT NULL DEFAULT '',
    stackable               BOOLEAN NOT NULL DEFAULT false,
    exclude_promotion_ids   UUID[] NOT NULL DEFAULT '{}',
    variant_ids             TEXT[] NOT NULL DEFAULT '{}',
    category_ids            TEXT[] NOT NULL DEFAULT '{}',
    brand_ids               TEXT[] NOT NULL DEFAULT '{}',
    segment_ids             TEXT[] NOT NULL DEFAULT '{}',
    global_limit            INT NOT NULL DEFAULT 0,
    per_user_limit          INT NOT NULL DEFAULT 0,
    per_order_limit         INT NOT NULL DEFAULT 0,
    per_device_limit        INT NOT NULL DEFAULT 0,
    min_qty                 INT NOT NULL DEFAULT 0,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);
