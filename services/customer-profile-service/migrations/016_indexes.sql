-- Performance indexes for profile, CRM, consent, segment, and privacy paths.

-- Profiles
CREATE INDEX idx_customer_profiles_tenant_status ON customer_profiles (tenant_id, status)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_customer_profiles_tenant_created ON customer_profiles (tenant_id, created_at DESC)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_customer_profiles_principal ON customer_profiles (principal_id);

-- Addresses
CREATE INDEX idx_addresses_profile ON addresses (profile_id)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_addresses_profile_default ON addresses (profile_id)
    WHERE is_default = TRUE AND deleted_at IS NULL;
CREATE INDEX idx_addresses_tenant_city ON addresses (tenant_id, city_id)
    WHERE deleted_at IS NULL AND city_id IS NOT NULL;
CREATE INDEX idx_addresses_geo ON addresses (lat, lng)
    WHERE deleted_at IS NULL;

-- Media
CREATE INDEX idx_profile_media_profile_current ON profile_media (profile_id)
    WHERE is_current = TRUE;

-- Tags
CREATE INDEX idx_tags_tenant_kind ON tags (tenant_id, kind);
CREATE INDEX idx_profile_tags_tag ON profile_tags (tag_id);
CREATE INDEX idx_profile_tags_assigned_at ON profile_tags (profile_id, assigned_at DESC);

-- Household
CREATE INDEX idx_households_tenant ON households (tenant_id)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_households_owner ON households (owner_profile_id)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_household_members_profile ON household_members (profile_id)
    WHERE left_at IS NULL;
CREATE INDEX idx_household_members_household ON household_members (household_id)
    WHERE left_at IS NULL;

-- Consents
CREATE INDEX idx_consents_tenant_channel ON consents (tenant_id, channel);
CREATE INDEX idx_consents_profile_granted ON consents (profile_id)
    WHERE granted = TRUE;

-- CRM
CREATE INDEX idx_crm_notes_profile_created ON crm_notes (profile_id, created_at DESC)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_crm_notes_tenant_created ON crm_notes (tenant_id, created_at DESC)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_timeline_events_profile_occurred ON timeline_events (profile_id, occurred_at DESC);
CREATE INDEX idx_timeline_events_tenant_type ON timeline_events (tenant_id, type, occurred_at DESC);
CREATE INDEX idx_timeline_events_actor ON timeline_events (actor_id, occurred_at DESC)
    WHERE actor_id IS NOT NULL;

-- Segments
CREATE INDEX idx_segments_tenant_kind ON segments (tenant_id, kind)
    WHERE active = TRUE;
CREATE INDEX idx_segment_members_profile ON segment_members (profile_id);
CREATE INDEX idx_segment_members_expires ON segment_members (expires_at)
    WHERE expires_at IS NOT NULL;

-- Personalization / AI (PK lookups are enough; light secondary)
CREATE INDEX idx_ai_customer_models_churn ON ai_customer_models (churn_prob DESC);
CREATE INDEX idx_ai_customer_models_updated ON ai_customer_models (updated_at DESC);

-- Loyalty / wallet projections
CREATE INDEX idx_loyalty_projections_tenant_level ON loyalty_projections (tenant_id, level);
CREATE INDEX idx_wallet_projections_tenant_status ON wallet_projections (tenant_id, status);

-- Activity
CREATE INDEX idx_profile_activity_profile_occurred ON profile_activity (profile_id, occurred_at DESC);
CREATE INDEX idx_profile_activity_tenant_occurred ON profile_activity (tenant_id, occurred_at DESC);
CREATE INDEX idx_profile_activity_actor ON profile_activity (actor_id, occurred_at DESC)
    WHERE actor_id IS NOT NULL;
CREATE INDEX idx_profile_activity_action ON profile_activity (action, occurred_at DESC);

-- Privacy
CREATE INDEX idx_privacy_requests_profile_created ON privacy_requests (profile_id, created_at DESC);
CREATE INDEX idx_privacy_requests_tenant_status ON privacy_requests (tenant_id, status, created_at DESC);
CREATE INDEX idx_privacy_requests_pending ON privacy_requests (kind, created_at)
    WHERE status IN ('pending', 'processing');

-- Merge jobs
CREATE INDEX idx_merge_jobs_tenant_status ON merge_jobs (tenant_id, status, created_at DESC);
CREATE INDEX idx_merge_jobs_source ON merge_jobs (source_profile_id);
CREATE INDEX idx_merge_jobs_target ON merge_jobs (target_profile_id);
