package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/location-service/internal/app/ports"
	"github.com/nexora/location-service/internal/domain"
)

// CacheRepo stores geocode cache and offline manifests.
type CacheRepo struct{ DB *sql.DB }

func (r *CacheRepo) GetGeocode(ctx context.Context, queryHash string) (domain.GeocodeResult, bool, error) {
	var result JSONGeocode
	var expiresAt time.Time
	err := r.DB.QueryRowContext(ctx, `
		SELECT result, expires_at FROM geocode_cache WHERE query_hash=$1`, queryHash).Scan(&result, &expiresAt)
	if isNoRows(err) {
		return domain.GeocodeResult{}, false, nil
	}
	if err != nil {
		return domain.GeocodeResult{}, false, err
	}
	if !expiresAt.IsZero() && time.Now().UTC().After(expiresAt.UTC()) {
		return domain.GeocodeResult{}, false, nil
	}
	out := domain.GeocodeResult(result)
	out.Cached = true
	return out, true, nil
}

func (r *CacheRepo) SetGeocode(ctx context.Context, queryHash string, result domain.GeocodeResult, expiresAt time.Time) error {
	payload := JSONGeocode(result)
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO geocode_cache (query_hash, result, expires_at, created_at)
		VALUES ($1,$2,$3,now())
		ON CONFLICT (query_hash) DO UPDATE SET
		  result=EXCLUDED.result, expires_at=EXCLUDED.expires_at`,
		queryHash, payload, expiresAt.UTC())
	return err
}

func (r *CacheRepo) GetOfflineManifest(ctx context.Context, tenantID uuid.UUID, region string) (domain.OfflineManifest, error) {
	var m domain.OfflineManifest
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, region, version, url, size_bytes, updated_at
		FROM offline_manifests WHERE tenant_id=$1 AND region=$2`, tenantID, region).Scan(
		&m.ID, &m.TenantID, &m.Region, &m.Version, &m.URL, &m.SizeBytes, &m.UpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.OfflineManifest{}, fmt.Errorf("%w: offline manifest", domain.ErrNotFound)
		}
		return domain.OfflineManifest{}, err
	}
	m.UpdatedAt = m.UpdatedAt.UTC()
	return m, nil
}

func (r *CacheRepo) UpsertOfflineManifest(ctx context.Context, m domain.OfflineManifest) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO offline_manifests (id, tenant_id, region, version, url, size_bytes, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (tenant_id, region) DO UPDATE SET
		  version=EXCLUDED.version, url=EXCLUDED.url, size_bytes=EXCLUDED.size_bytes,
		  updated_at=EXCLUDED.updated_at`,
		m.ID, m.TenantID, m.Region, m.Version, m.URL, m.SizeBytes, m.UpdatedAt.UTC())
	return err
}

var _ ports.CacheRepo = (*CacheRepo)(nil)
