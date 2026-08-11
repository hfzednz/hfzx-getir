package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/nexora/global-service/internal/app/ports"
	"github.com/nexora/global-service/internal/domain"
)

type HolidayRepo struct{ DB *sql.DB }

func (r *HolidayRepo) Save(ctx context.Context, h domain.Holiday) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO global_holidays (
			id, tenant_id, country_id, date, name, kind, created_at
		) VALUES ($1,$2,$3,$4::date,$5,$6,$7)
		ON CONFLICT (id) DO UPDATE SET
			date=EXCLUDED.date, name=EXCLUDED.name, kind=EXCLUDED.kind`,
		h.ID, h.TenantID, h.CountryID, dateOnly(h.Date), h.Name, h.Kind, h.CreatedAt.UTC())
	return err
}

func (r *HolidayRepo) ListByCountry(ctx context.Context, tenantID, countryID uuid.UUID) ([]domain.Holiday, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, country_id, date, name, kind, created_at
		FROM global_holidays WHERE tenant_id=$1 AND country_id=$2 ORDER BY date ASC`, tenantID, countryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Holiday{}
	for rows.Next() {
		var h domain.Holiday
		var d sql.NullTime
		if err := rows.Scan(&h.ID, &h.TenantID, &h.CountryID, &d, &h.Name, &h.Kind, &h.CreatedAt); err != nil {
			return nil, err
		}
		h.Date = scanDateString(d)
		h.CreatedAt = h.CreatedAt.UTC()
		out = append(out, h)
	}
	return out, rows.Err()
}

type RuleRepo struct{ DB *sql.DB }

func (r *RuleRepo) Save(ctx context.Context, rule domain.RegionalRule) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO global_regional_rules (
			id, tenant_id, country_id, place_id, min_order_minor, delivery_fee_minor, currency,
			legal_age, restricted_skus, open_hour, close_hour, warehouse_rules, courier_rules,
			active, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		ON CONFLICT (id) DO UPDATE SET
			place_id=EXCLUDED.place_id, min_order_minor=EXCLUDED.min_order_minor,
			delivery_fee_minor=EXCLUDED.delivery_fee_minor, currency=EXCLUDED.currency,
			legal_age=EXCLUDED.legal_age, restricted_skus=EXCLUDED.restricted_skus,
			open_hour=EXCLUDED.open_hour, close_hour=EXCLUDED.close_hour,
			warehouse_rules=EXCLUDED.warehouse_rules, courier_rules=EXCLUDED.courier_rules,
			active=EXCLUDED.active, updated_at=EXCLUDED.updated_at`,
		rule.ID, rule.TenantID, rule.CountryID, nullUUID(rule.PlaceID), rule.MinOrderMinor, rule.DeliveryFeeMinor,
		rule.Currency, rule.LegalAge, textArray(rule.RestrictedSKUs), rule.OpenHour, rule.CloseHour,
		JSONMap(rule.WarehouseRules), JSONMap(rule.CourierRules), rule.Active, rule.UpdatedAt.UTC())
	return err
}

func (r *RuleRepo) GetForCountry(ctx context.Context, tenantID, countryID uuid.UUID) (domain.RegionalRule, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, country_id, place_id, min_order_minor, delivery_fee_minor, currency,
			legal_age, restricted_skus, open_hour, close_hour, warehouse_rules, courier_rules,
			active, updated_at
		FROM global_regional_rules WHERE tenant_id=$1 AND country_id=$2`, tenantID, countryID)
	var rule domain.RegionalRule
	var place uuid.NullUUID
	var skus []string
	var wh, courier JSONMap
	err := row.Scan(&rule.ID, &rule.TenantID, &rule.CountryID, &place, &rule.MinOrderMinor, &rule.DeliveryFeeMinor,
		&rule.Currency, &rule.LegalAge, pq.Array(&skus), &rule.OpenHour, &rule.CloseHour, &wh, &courier,
		&rule.Active, &rule.UpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.RegionalRule{}, domain.ErrNotFound
		}
		return domain.RegionalRule{}, err
	}
	rule.PlaceID = scanNullUUID(place)
	rule.RestrictedSKUs = skus
	if rule.RestrictedSKUs == nil {
		rule.RestrictedSKUs = []string{}
	}
	rule.WarehouseRules = map[string]any(wh)
	rule.CourierRules = map[string]any(courier)
	rule.UpdatedAt = rule.UpdatedAt.UTC()
	return rule, nil
}

type TaxRepo struct{ DB *sql.DB }

func (r *TaxRepo) Save(ctx context.Context, t domain.TaxRule) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO global_tax_rules (
			id, tenant_id, country_id, place_id, kind, rate_bps, name, exempt_skus, active, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (id) DO UPDATE SET
			place_id=EXCLUDED.place_id, kind=EXCLUDED.kind, rate_bps=EXCLUDED.rate_bps,
			name=EXCLUDED.name, exempt_skus=EXCLUDED.exempt_skus, active=EXCLUDED.active,
			updated_at=EXCLUDED.updated_at`,
		t.ID, t.TenantID, t.CountryID, nullUUID(t.PlaceID), t.Kind, t.RateBps, t.Name,
		textArray(t.ExemptSKUs), t.Active, t.UpdatedAt.UTC())
	return err
}

func (r *TaxRepo) ListByCountry(ctx context.Context, tenantID, countryID uuid.UUID) ([]domain.TaxRule, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, country_id, place_id, kind, rate_bps, name, exempt_skus, active, updated_at
		FROM global_tax_rules WHERE tenant_id=$1 AND country_id=$2 ORDER BY name ASC`, tenantID, countryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.TaxRule{}
	for rows.Next() {
		var t domain.TaxRule
		var place uuid.NullUUID
		var skus []string
		if err := rows.Scan(&t.ID, &t.TenantID, &t.CountryID, &place, &t.Kind, &t.RateBps, &t.Name,
			pq.Array(&skus), &t.Active, &t.UpdatedAt); err != nil {
			return nil, err
		}
		t.PlaceID = scanNullUUID(place)
		t.ExemptSKUs = skus
		if t.ExemptSKUs == nil {
			t.ExemptSKUs = []string{}
		}
		t.UpdatedAt = t.UpdatedAt.UTC()
		out = append(out, t)
	}
	return out, rows.Err()
}

type PrivacyRepo struct{ DB *sql.DB }

func (r *PrivacyRepo) Save(ctx context.Context, p domain.PrivacyRegime) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO global_privacy (
			id, tenant_id, country_id, framework, consent_required, retention_days, data_residency, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (id) DO UPDATE SET
			framework=EXCLUDED.framework, consent_required=EXCLUDED.consent_required,
			retention_days=EXCLUDED.retention_days, data_residency=EXCLUDED.data_residency,
			updated_at=EXCLUDED.updated_at`,
		p.ID, p.TenantID, p.CountryID, p.Framework, p.ConsentRequired, p.RetentionDays, p.DataResidency, p.UpdatedAt.UTC())
	return err
}

func (r *PrivacyRepo) GetByCountry(ctx context.Context, tenantID, countryID uuid.UUID) (domain.PrivacyRegime, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, country_id, framework, consent_required, retention_days, data_residency, updated_at
		FROM global_privacy WHERE tenant_id=$1 AND country_id=$2`, tenantID, countryID)
	var p domain.PrivacyRegime
	err := row.Scan(&p.ID, &p.TenantID, &p.CountryID, &p.Framework, &p.ConsentRequired, &p.RetentionDays, &p.DataResidency, &p.UpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.PrivacyRegime{}, domain.ErrNotFound
		}
		return domain.PrivacyRegime{}, err
	}
	p.UpdatedAt = p.UpdatedAt.UTC()
	return p, nil
}

type PayAvailRepo struct{ DB *sql.DB }

func (r *PayAvailRepo) Save(ctx context.Context, p domain.PaymentMethodAvailability) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO global_payment_availability (
			id, tenant_id, country_id, method_code, enabled, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (id) DO UPDATE SET
			method_code=EXCLUDED.method_code, enabled=EXCLUDED.enabled, updated_at=EXCLUDED.updated_at`,
		p.ID, p.TenantID, p.CountryID, p.MethodCode, p.Enabled, p.UpdatedAt.UTC())
	return err
}

func (r *PayAvailRepo) ListByCountry(ctx context.Context, tenantID, countryID uuid.UUID) ([]domain.PaymentMethodAvailability, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, country_id, method_code, enabled, updated_at
		FROM global_payment_availability WHERE tenant_id=$1 AND country_id=$2 ORDER BY method_code ASC`, tenantID, countryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.PaymentMethodAvailability{}
	for rows.Next() {
		var p domain.PaymentMethodAvailability
		if err := rows.Scan(&p.ID, &p.TenantID, &p.CountryID, &p.MethodCode, &p.Enabled, &p.UpdatedAt); err != nil {
			return nil, err
		}
		p.UpdatedAt = p.UpdatedAt.UTC()
		out = append(out, p)
	}
	return out, rows.Err()
}

type LogisticsRepo struct{ DB *sql.DB }

func (r *LogisticsRepo) Save(ctx context.Context, p domain.LogisticsPolicy) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO global_logistics_policy (
			id, tenant_id, country_id, sla_minutes, holiday_routing, zone_codes, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (id) DO UPDATE SET
			sla_minutes=EXCLUDED.sla_minutes, holiday_routing=EXCLUDED.holiday_routing,
			zone_codes=EXCLUDED.zone_codes, updated_at=EXCLUDED.updated_at`,
		p.ID, p.TenantID, p.CountryID, p.SLAMinutes, p.HolidayRouting, textArray(p.ZoneCodes), p.UpdatedAt.UTC())
	return err
}

func (r *LogisticsRepo) GetByCountry(ctx context.Context, tenantID, countryID uuid.UUID) (domain.LogisticsPolicy, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, country_id, sla_minutes, holiday_routing, zone_codes, updated_at
		FROM global_logistics_policy WHERE tenant_id=$1 AND country_id=$2`, tenantID, countryID)
	var p domain.LogisticsPolicy
	var zones []string
	err := row.Scan(&p.ID, &p.TenantID, &p.CountryID, &p.SLAMinutes, &p.HolidayRouting, pq.Array(&zones), &p.UpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.LogisticsPolicy{}, domain.ErrNotFound
		}
		return domain.LogisticsPolicy{}, err
	}
	p.ZoneCodes = zones
	if p.ZoneCodes == nil {
		p.ZoneCodes = []string{}
	}
	p.UpdatedAt = p.UpdatedAt.UTC()
	return p, nil
}

type LegalRepo struct{ DB *sql.DB }

func (r *LegalRepo) Save(ctx context.Context, d domain.LegalDocument) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO global_legal_docs (
			id, tenant_id, country_id, kind, locale, version, uri, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (id) DO UPDATE SET
			kind=EXCLUDED.kind, locale=EXCLUDED.locale, version=EXCLUDED.version,
			uri=EXCLUDED.uri, updated_at=EXCLUDED.updated_at`,
		d.ID, d.TenantID, d.CountryID, d.Kind, d.Locale, d.Version, d.URI, d.UpdatedAt.UTC())
	return err
}

func (r *LegalRepo) List(ctx context.Context, tenantID, countryID uuid.UUID, locale string) ([]domain.LegalDocument, error) {
	var (
		rows *sql.Rows
		err  error
	)
	if locale == "" {
		rows, err = r.DB.QueryContext(ctx, `
			SELECT id, tenant_id, country_id, kind, locale, version, uri, updated_at
			FROM global_legal_docs WHERE tenant_id=$1 AND country_id=$2 ORDER BY kind ASC, locale ASC`,
			tenantID, countryID)
	} else {
		rows, err = r.DB.QueryContext(ctx, `
			SELECT id, tenant_id, country_id, kind, locale, version, uri, updated_at
			FROM global_legal_docs WHERE tenant_id=$1 AND country_id=$2 AND locale=$3 ORDER BY kind ASC`,
			tenantID, countryID, locale)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.LegalDocument{}
	for rows.Next() {
		var d domain.LegalDocument
		if err := rows.Scan(&d.ID, &d.TenantID, &d.CountryID, &d.Kind, &d.Locale, &d.Version, &d.URI, &d.UpdatedAt); err != nil {
			return nil, err
		}
		d.UpdatedAt = d.UpdatedAt.UTC()
		out = append(out, d)
	}
	return out, rows.Err()
}

var (
	_ ports.HolidayRepo  = (*HolidayRepo)(nil)
	_ ports.RuleRepo     = (*RuleRepo)(nil)
	_ ports.TaxRepo      = (*TaxRepo)(nil)
	_ ports.PrivacyRepo  = (*PrivacyRepo)(nil)
	_ ports.PayAvailRepo = (*PayAvailRepo)(nil)
	_ ports.LogisticsRepo = (*LogisticsRepo)(nil)
	_ ports.LegalRepo    = (*LegalRepo)(nil)
)
