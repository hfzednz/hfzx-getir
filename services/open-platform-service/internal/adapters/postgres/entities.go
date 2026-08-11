package postgres

import (
	"context"
	"database/sql"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/open-platform-service/internal/app/ports"
	"github.com/nexora/open-platform-service/internal/domain"
)

type scannable interface{ Scan(dest ...any) error }

type AppRepo struct{ DB *sql.DB }

func (r *AppRepo) Save(ctx context.Context, a domain.DeveloperApp) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO open_apps (id, tenant_id, name, kind, owner_email, scopes, oauth_client_id, sandbox, active, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (id) DO UPDATE SET name=EXCLUDED.name, kind=EXCLUDED.kind, owner_email=EXCLUDED.owner_email,
			scopes=EXCLUDED.scopes, oauth_client_id=EXCLUDED.oauth_client_id, sandbox=EXCLUDED.sandbox, active=EXCLUDED.active`,
		a.ID, a.TenantID, a.Name, string(a.Kind), a.OwnerEmail, TextArray(a.Scopes), a.OAuthClientID, a.Sandbox, a.Active, a.CreatedAt.UTC())
	return err
}

func (r *AppRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.DeveloperApp, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, name, kind, owner_email, scopes, oauth_client_id, sandbox, active, created_at
		FROM open_apps WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	var a domain.DeveloperApp
	var kind string
	var scopes TextArray
	err := row.Scan(&a.ID, &a.TenantID, &a.Name, &kind, &a.OwnerEmail, &scopes, &a.OAuthClientID, &a.Sandbox, &a.Active, &a.CreatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.DeveloperApp{}, domain.ErrNotFound
		}
		return domain.DeveloperApp{}, err
	}
	a.Kind = domain.AppKind(kind)
	a.Scopes = []string(scopes)
	a.CreatedAt = a.CreatedAt.UTC()
	return a, nil
}

func (r *AppRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.DeveloperApp, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, name, kind, owner_email, scopes, oauth_client_id, sandbox, active, created_at
		FROM open_apps WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.DeveloperApp{}
	for rows.Next() {
		var a domain.DeveloperApp
		var kind string
		var scopes TextArray
		if err := rows.Scan(&a.ID, &a.TenantID, &a.Name, &kind, &a.OwnerEmail, &scopes, &a.OAuthClientID, &a.Sandbox, &a.Active, &a.CreatedAt); err != nil {
			return nil, err
		}
		a.Kind = domain.AppKind(kind)
		a.Scopes = []string(scopes)
		a.CreatedAt = a.CreatedAt.UTC()
		out = append(out, a)
	}
	return out, rows.Err()
}

type ApiKeyRepo struct{ DB *sql.DB }

func (r *ApiKeyRepo) Save(ctx context.Context, k domain.ApiKey) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO open_api_keys (id, tenant_id, app_id, name, prefix, secret_hash, scopes, expires_at, created_at, revoked)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (id) DO UPDATE SET name=EXCLUDED.name, revoked=EXCLUDED.revoked, scopes=EXCLUDED.scopes, expires_at=EXCLUDED.expires_at`,
		k.ID, k.TenantID, k.AppID, k.Name, k.Prefix, k.SecretHash, TextArray(k.Scopes), nullTime(k.ExpiresAt), k.CreatedAt.UTC(), k.Revoked)
	return err
}

func (r *ApiKeyRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.ApiKey, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, app_id, name, prefix, secret_hash, scopes, expires_at, created_at, revoked
		FROM open_api_keys WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	return scanKey(row)
}

func (r *ApiKeyRepo) ListByApp(ctx context.Context, tenantID, appID uuid.UUID) ([]domain.ApiKey, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, app_id, name, prefix, secret_hash, scopes, expires_at, created_at, revoked
		FROM open_api_keys WHERE tenant_id=$1 AND app_id=$2 ORDER BY created_at DESC`, tenantID, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ApiKey{}
	for rows.Next() {
		k, err := scanKey(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, k)
	}
	return out, rows.Err()
}

func scanKey(row scannable) (domain.ApiKey, error) {
	var k domain.ApiKey
	var scopes TextArray
	var exp sql.NullTime
	err := row.Scan(&k.ID, &k.TenantID, &k.AppID, &k.Name, &k.Prefix, &k.SecretHash, &scopes, &exp, &k.CreatedAt, &k.Revoked)
	if err != nil {
		if isNoRows(err) {
			return domain.ApiKey{}, domain.ErrNotFound
		}
		return domain.ApiKey{}, err
	}
	k.Scopes = []string(scopes)
	k.ExpiresAt = scanNullTime(exp)
	k.CreatedAt = k.CreatedAt.UTC()
	return k, nil
}

type CatalogRepo struct{ DB *sql.DB }

func (r *CatalogRepo) Save(ctx context.Context, e domain.CatalogEntry) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO open_catalog (id, tenant_id, key, title, surface, base_path, service_ref, openapi_path, active, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (tenant_id, key) DO UPDATE SET id=EXCLUDED.id, title=EXCLUDED.title, surface=EXCLUDED.surface,
			base_path=EXCLUDED.base_path, service_ref=EXCLUDED.service_ref, openapi_path=EXCLUDED.openapi_path, active=EXCLUDED.active`,
		e.ID, e.TenantID, e.Key, e.Title, string(e.Surface), e.BasePath, e.ServiceRef, e.OpenAPIPath, e.Active, e.CreatedAt.UTC())
	return err
}

func (r *CatalogRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.CatalogEntry, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, key, title, surface, base_path, service_ref, openapi_path, active, created_at
		FROM open_catalog WHERE tenant_id=$1 ORDER BY key ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.CatalogEntry{}
	for rows.Next() {
		var e domain.CatalogEntry
		var surface string
		if err := rows.Scan(&e.ID, &e.TenantID, &e.Key, &e.Title, &surface, &e.BasePath, &e.ServiceRef, &e.OpenAPIPath, &e.Active, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.Surface = domain.Surface(surface)
		e.CreatedAt = e.CreatedAt.UTC()
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *CatalogRepo) GetByKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.CatalogEntry, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, key, title, surface, base_path, service_ref, openapi_path, active, created_at
		FROM open_catalog WHERE tenant_id=$1 AND key=$2`, tenantID, key)
	var e domain.CatalogEntry
	var surface string
	err := row.Scan(&e.ID, &e.TenantID, &e.Key, &e.Title, &surface, &e.BasePath, &e.ServiceRef, &e.OpenAPIPath, &e.Active, &e.CreatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.CatalogEntry{}, domain.ErrNotFound
		}
		return domain.CatalogEntry{}, err
	}
	e.Surface = domain.Surface(surface)
	e.CreatedAt = e.CreatedAt.UTC()
	return e, nil
}

type VersionRepo struct{ DB *sql.DB }

func (r *VersionRepo) Save(ctx context.Context, v domain.ApiVersion) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO open_api_versions (id, tenant_id, catalog_key, version, status, released_at, deprecated_at, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (id) DO UPDATE SET status=EXCLUDED.status, deprecated_at=EXCLUDED.deprecated_at, notes=EXCLUDED.notes`,
		v.ID, v.TenantID, v.CatalogKey, v.Version, v.Status, v.ReleasedAt.UTC(), nullTime(v.DeprecatedAt), v.Notes)
	return err
}

func (r *VersionRepo) ListByCatalog(ctx context.Context, tenantID uuid.UUID, catalogKey string) ([]domain.ApiVersion, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, catalog_key, version, status, released_at, deprecated_at, notes
		FROM open_api_versions WHERE tenant_id=$1 AND catalog_key=$2 ORDER BY released_at DESC`, tenantID, catalogKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ApiVersion{}
	for rows.Next() {
		var v domain.ApiVersion
		var dep sql.NullTime
		if err := rows.Scan(&v.ID, &v.TenantID, &v.CatalogKey, &v.Version, &v.Status, &v.ReleasedAt, &dep, &v.Notes); err != nil {
			return nil, err
		}
		v.DeprecatedAt = scanNullTime(dep)
		v.ReleasedAt = v.ReleasedAt.UTC()
		out = append(out, v)
	}
	return out, rows.Err()
}

type PolicyRepo struct{ DB *sql.DB }

func (r *PolicyRepo) Save(ctx context.Context, p domain.GatewayPolicy) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO open_gateway_policies (
			id, tenant_id, name, route_prefix, target_service, version, rate_limit_rpm, quota_daily,
			canary_percent, require_scopes, cache_ttl_seconds, active, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		ON CONFLICT (id) DO UPDATE SET name=EXCLUDED.name, route_prefix=EXCLUDED.route_prefix,
			target_service=EXCLUDED.target_service, version=EXCLUDED.version, rate_limit_rpm=EXCLUDED.rate_limit_rpm,
			quota_daily=EXCLUDED.quota_daily, canary_percent=EXCLUDED.canary_percent, require_scopes=EXCLUDED.require_scopes,
			cache_ttl_seconds=EXCLUDED.cache_ttl_seconds, active=EXCLUDED.active, updated_at=EXCLUDED.updated_at`,
		p.ID, p.TenantID, p.Name, p.RoutePrefix, p.TargetService, p.Version, p.RateLimitRPM, p.QuotaDaily,
		p.CanaryPercent, TextArray(p.RequireScopes), p.CacheTTLSeconds, p.Active, p.UpdatedAt.UTC())
	return err
}

func (r *PolicyRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.GatewayPolicy, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, name, route_prefix, target_service, version, rate_limit_rpm, quota_daily,
			canary_percent, require_scopes, cache_ttl_seconds, active, updated_at
		FROM open_gateway_policies WHERE tenant_id=$1 ORDER BY name ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.GatewayPolicy{}
	for rows.Next() {
		var p domain.GatewayPolicy
		var scopes TextArray
		if err := rows.Scan(&p.ID, &p.TenantID, &p.Name, &p.RoutePrefix, &p.TargetService, &p.Version, &p.RateLimitRPM, &p.QuotaDaily,
			&p.CanaryPercent, &scopes, &p.CacheTTLSeconds, &p.Active, &p.UpdatedAt); err != nil {
			return nil, err
		}
		p.RequireScopes = []string(scopes)
		p.UpdatedAt = p.UpdatedAt.UTC()
		out = append(out, p)
	}
	return out, rows.Err()
}

type WebhookRepo struct{ DB *sql.DB }

func (r *WebhookRepo) Save(ctx context.Context, w domain.WebhookEndpoint) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO open_webhooks (id, tenant_id, app_id, url, secret, events, version, active, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (id) DO UPDATE SET url=EXCLUDED.url, secret=EXCLUDED.secret, events=EXCLUDED.events,
			version=EXCLUDED.version, active=EXCLUDED.active`,
		w.ID, w.TenantID, w.AppID, w.URL, w.Secret, TextArray(w.Events), w.Version, w.Active, w.CreatedAt.UTC())
	return err
}

func (r *WebhookRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.WebhookEndpoint, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, app_id, url, secret, events, version, active, created_at
		FROM open_webhooks WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	return scanWebhook(row)
}

func (r *WebhookRepo) ListByApp(ctx context.Context, tenantID, appID uuid.UUID) ([]domain.WebhookEndpoint, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, app_id, url, secret, events, version, active, created_at
		FROM open_webhooks WHERE tenant_id=$1 AND app_id=$2`, tenantID, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.WebhookEndpoint{}
	for rows.Next() {
		w, err := scanWebhook(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

func (r *WebhookRepo) ListActiveForEvent(ctx context.Context, tenantID uuid.UUID, eventType string) ([]domain.WebhookEndpoint, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, app_id, url, secret, events, version, active, created_at
		FROM open_webhooks WHERE tenant_id=$1 AND active=TRUE`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.WebhookEndpoint{}
	for rows.Next() {
		w, err := scanWebhook(rows)
		if err != nil {
			return nil, err
		}
		for _, ev := range w.Events {
			if ev == eventType || ev == "*" {
				out = append(out, w)
				break
			}
		}
	}
	return out, rows.Err()
}

func scanWebhook(row scannable) (domain.WebhookEndpoint, error) {
	var w domain.WebhookEndpoint
	var events TextArray
	err := row.Scan(&w.ID, &w.TenantID, &w.AppID, &w.URL, &w.Secret, &events, &w.Version, &w.Active, &w.CreatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.WebhookEndpoint{}, domain.ErrNotFound
		}
		return domain.WebhookEndpoint{}, err
	}
	w.Events = []string(events)
	w.CreatedAt = w.CreatedAt.UTC()
	return w, nil
}

type DeliveryRepo struct{ DB *sql.DB }

func (r *DeliveryRepo) Save(ctx context.Context, d domain.WebhookDelivery) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO open_webhook_deliveries (
			id, tenant_id, endpoint_id, event_type, payload, status, attempts, last_http_status, last_error, signature, created_at, delivered_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (id) DO UPDATE SET status=EXCLUDED.status, attempts=EXCLUDED.attempts,
			last_http_status=EXCLUDED.last_http_status, last_error=EXCLUDED.last_error, delivered_at=EXCLUDED.delivered_at`,
		d.ID, d.TenantID, d.EndpointID, d.EventType, JSONMap(d.Payload), string(d.Status), d.Attempts, d.LastStatus, d.LastError, d.Signature, d.CreatedAt.UTC(), nullTime(d.DeliveredAt))
	return err
}

func (r *DeliveryRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.WebhookDelivery, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, endpoint_id, event_type, payload, status, attempts, last_http_status, last_error, signature, created_at, delivered_at
		FROM open_webhook_deliveries WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	return scanDelivery(row)
}

func (r *DeliveryRepo) ListPending(ctx context.Context, limit int) ([]domain.WebhookDelivery, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, endpoint_id, event_type, payload, status, attempts, last_http_status, last_error, signature, created_at, delivered_at
		FROM open_webhook_deliveries WHERE status=$1 ORDER BY created_at ASC LIMIT $2`, string(domain.DeliveryPending), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDeliveries(rows)
}

func (r *DeliveryRepo) ListByEndpoint(ctx context.Context, tenantID, endpointID uuid.UUID) ([]domain.WebhookDelivery, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, endpoint_id, event_type, payload, status, attempts, last_http_status, last_error, signature, created_at, delivered_at
		FROM open_webhook_deliveries WHERE tenant_id=$1 AND endpoint_id=$2 ORDER BY created_at DESC`, tenantID, endpointID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDeliveries(rows)
}

func scanDeliveries(rows *sql.Rows) ([]domain.WebhookDelivery, error) {
	out := []domain.WebhookDelivery{}
	for rows.Next() {
		d, err := scanDelivery(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func scanDelivery(row scannable) (domain.WebhookDelivery, error) {
	var d domain.WebhookDelivery
	var status string
	var payload JSONMap
	var delivered sql.NullTime
	err := row.Scan(&d.ID, &d.TenantID, &d.EndpointID, &d.EventType, &payload, &status, &d.Attempts, &d.LastStatus, &d.LastError, &d.Signature, &d.CreatedAt, &delivered)
	if err != nil {
		if isNoRows(err) {
			return domain.WebhookDelivery{}, domain.ErrNotFound
		}
		return domain.WebhookDelivery{}, err
	}
	d.Payload = map[string]any(payload)
	d.Status = domain.DeliveryStatus(status)
	d.DeliveredAt = scanNullTime(delivered)
	d.CreatedAt = d.CreatedAt.UTC()
	return d, nil
}

type SdkRepo struct{ DB *sql.DB }

func (r *SdkRepo) Save(ctx context.Context, s domain.SdkPackage) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO open_sdks (id, tenant_id, language, name, version, repo_path, status, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (id) DO UPDATE SET status=EXCLUDED.status, version=EXCLUDED.version, repo_path=EXCLUDED.repo_path`,
		s.ID, s.TenantID, string(s.Language), s.Name, s.Version, s.RepoPath, s.Status, s.CreatedAt.UTC())
	return err
}

func (r *SdkRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.SdkPackage, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, language, name, version, repo_path, status, created_at
		FROM open_sdks WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.SdkPackage{}
	for rows.Next() {
		var s domain.SdkPackage
		var lang string
		if err := rows.Scan(&s.ID, &s.TenantID, &lang, &s.Name, &s.Version, &s.RepoPath, &s.Status, &s.CreatedAt); err != nil {
			return nil, err
		}
		s.Language = domain.SdkLanguage(lang)
		s.CreatedAt = s.CreatedAt.UTC()
		out = append(out, s)
	}
	return out, rows.Err()
}

type IntegrationRepo struct{ DB *sql.DB }

func (r *IntegrationRepo) Save(ctx context.Context, i domain.IntegrationConnector) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO open_integrations (id, tenant_id, app_id, kind, provider, config_ref, status, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (id) DO UPDATE SET status=EXCLUDED.status, config_ref=EXCLUDED.config_ref, provider=EXCLUDED.provider`,
		i.ID, i.TenantID, i.AppID, string(i.Kind), i.Provider, i.ConfigRef, i.Status, i.CreatedAt.UTC())
	return err
}

func (r *IntegrationRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.IntegrationConnector, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, app_id, kind, provider, config_ref, status, created_at
		FROM open_integrations WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.IntegrationConnector{}
	for rows.Next() {
		var i domain.IntegrationConnector
		var kind string
		if err := rows.Scan(&i.ID, &i.TenantID, &i.AppID, &kind, &i.Provider, &i.ConfigRef, &i.Status, &i.CreatedAt); err != nil {
			return nil, err
		}
		i.Kind = domain.IntegrationKind(kind)
		i.CreatedAt = i.CreatedAt.UTC()
		out = append(out, i)
	}
	return out, rows.Err()
}

type UsageRepo struct{ DB *sql.DB }

func (r *UsageRepo) Save(ctx context.Context, u domain.UsageCounter) error {
	day := strings.TrimSpace(u.Day)
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO open_usage (id, tenant_id, app_id, day, requests, errors, updated_at)
		VALUES ($1,$2,$3,$4::date,$5,$6,$7)
		ON CONFLICT (tenant_id, app_id, day) DO UPDATE SET
			id=EXCLUDED.id, requests=EXCLUDED.requests, errors=EXCLUDED.errors, updated_at=EXCLUDED.updated_at`,
		u.ID, u.TenantID, u.AppID, day, u.Requests, u.Errors, u.UpdatedAt.UTC())
	return err
}

func (r *UsageRepo) Get(ctx context.Context, tenantID, appID uuid.UUID, day string) (domain.UsageCounter, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, app_id, to_char(day,'YYYY-MM-DD'), requests, errors, updated_at
		FROM open_usage WHERE tenant_id=$1 AND app_id=$2 AND day=$3::date`, tenantID, appID, day)
	var u domain.UsageCounter
	err := row.Scan(&u.ID, &u.TenantID, &u.AppID, &u.Day, &u.Requests, &u.Errors, &u.UpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.UsageCounter{}, domain.ErrNotFound
		}
		return domain.UsageCounter{}, err
	}
	u.UpdatedAt = u.UpdatedAt.UTC()
	return u, nil
}

func (r *UsageRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.UsageCounter, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, app_id, to_char(day,'YYYY-MM-DD'), requests, errors, updated_at
		FROM open_usage WHERE tenant_id=$1 ORDER BY day DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.UsageCounter{}
	for rows.Next() {
		var u domain.UsageCounter
		if err := rows.Scan(&u.ID, &u.TenantID, &u.AppID, &u.Day, &u.Requests, &u.Errors, &u.UpdatedAt); err != nil {
			return nil, err
		}
		u.UpdatedAt = u.UpdatedAt.UTC()
		out = append(out, u)
	}
	return out, rows.Err()
}

var (
	_ ports.AppRepo         = (*AppRepo)(nil)
	_ ports.ApiKeyRepo      = (*ApiKeyRepo)(nil)
	_ ports.CatalogRepo     = (*CatalogRepo)(nil)
	_ ports.VersionRepo     = (*VersionRepo)(nil)
	_ ports.PolicyRepo      = (*PolicyRepo)(nil)
	_ ports.WebhookRepo     = (*WebhookRepo)(nil)
	_ ports.DeliveryRepo    = (*DeliveryRepo)(nil)
	_ ports.SdkRepo         = (*SdkRepo)(nil)
	_ ports.IntegrationRepo = (*IntegrationRepo)(nil)
	_ ports.UsageRepo       = (*UsageRepo)(nil)
)
