package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/location-service/internal/app/ports"
	"github.com/nexora/location-service/internal/domain"
)

// AddressRepo persists normalized addresses.
type AddressRepo struct{ DB *sql.DB }

func (r *AddressRepo) Upsert(ctx context.Context, a domain.NormalizedAddress) error {
	comp := JSONComponents(a.Components)
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO addresses
		  (id, tenant_id, line1, building, entrance, floor, apt, landmark, place_id,
		   lat, lng, confidence, risk_score, components, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
		ON CONFLICT (id) DO UPDATE SET
		  tenant_id=EXCLUDED.tenant_id, line1=EXCLUDED.line1, building=EXCLUDED.building,
		  entrance=EXCLUDED.entrance, floor=EXCLUDED.floor, apt=EXCLUDED.apt,
		  landmark=EXCLUDED.landmark, place_id=EXCLUDED.place_id, lat=EXCLUDED.lat, lng=EXCLUDED.lng,
		  confidence=EXCLUDED.confidence, risk_score=EXCLUDED.risk_score,
		  components=EXCLUDED.components, updated_at=EXCLUDED.updated_at`,
		a.ID, a.TenantID, a.Line1, a.Building, a.Entrance, a.Floor, a.Apt, a.Landmark, a.PlaceID,
		a.Lat, a.Lng, float64(a.Confidence), a.RiskScore, comp, a.CreatedAt.UTC(), a.UpdatedAt.UTC())
	return err
}

func (r *AddressRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.NormalizedAddress, error) {
	var a domain.NormalizedAddress
	var conf float64
	var comp JSONComponents
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, line1, building, entrance, floor, apt, landmark, place_id,
		       lat, lng, confidence, risk_score, components, created_at, updated_at
		FROM addresses WHERE id=$1 AND tenant_id=$2`, id, tenantID).Scan(
		&a.ID, &a.TenantID, &a.Line1, &a.Building, &a.Entrance, &a.Floor, &a.Apt, &a.Landmark, &a.PlaceID,
		&a.Lat, &a.Lng, &conf, &a.RiskScore, &comp, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.NormalizedAddress{}, fmt.Errorf("%w: address", domain.ErrNotFound)
		}
		return domain.NormalizedAddress{}, err
	}
	a.Confidence = domain.ConfidenceScore(conf)
	a.Components = domain.AddressComponents(comp)
	a.CreatedAt = a.CreatedAt.UTC()
	a.UpdatedAt = a.UpdatedAt.UTC()
	return a, nil
}

var _ ports.AddressRepo = (*AddressRepo)(nil)
