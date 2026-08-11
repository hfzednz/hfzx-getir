package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/global-service/internal/app/ports"
	"github.com/nexora/global-service/internal/domain"
)

type CountryRepo struct{ DB *sql.DB }

func (r *CountryRepo) Save(ctx context.Context, c domain.Country) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO global_countries (
			id, tenant_id, code, name, default_locale, default_currency, default_tz, status,
			data_residency, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (id) DO UPDATE SET
			code=EXCLUDED.code, name=EXCLUDED.name, default_locale=EXCLUDED.default_locale,
			default_currency=EXCLUDED.default_currency, default_tz=EXCLUDED.default_tz,
			status=EXCLUDED.status, data_residency=EXCLUDED.data_residency, updated_at=EXCLUDED.updated_at`,
		c.ID, c.TenantID, c.Code, c.Name, c.DefaultLocale, c.DefaultCurrency, c.DefaultTZ, c.Status,
		c.DataResidency, c.CreatedAt.UTC(), c.UpdatedAt.UTC())
	return err
}

func (r *CountryRepo) scan(row interface{ Scan(dest ...any) error }) (domain.Country, error) {
	var c domain.Country
	err := row.Scan(&c.ID, &c.TenantID, &c.Code, &c.Name, &c.DefaultLocale, &c.DefaultCurrency, &c.DefaultTZ,
		&c.Status, &c.DataResidency, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return domain.Country{}, err
	}
	c.CreatedAt = c.CreatedAt.UTC()
	c.UpdatedAt = c.UpdatedAt.UTC()
	return c, nil
}

func (r *CountryRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Country, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, code, name, default_locale, default_currency, default_tz, status,
			data_residency, created_at, updated_at
		FROM global_countries WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	c, err := r.scan(row)
	if err != nil {
		if isNoRows(err) {
			return domain.Country{}, domain.ErrNotFound
		}
		return domain.Country{}, err
	}
	return c, nil
}

func (r *CountryRepo) GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (domain.Country, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, code, name, default_locale, default_currency, default_tz, status,
			data_residency, created_at, updated_at
		FROM global_countries WHERE tenant_id=$1 AND code=$2`, tenantID, code)
	c, err := r.scan(row)
	if err != nil {
		if isNoRows(err) {
			return domain.Country{}, domain.ErrNotFound
		}
		return domain.Country{}, err
	}
	return c, nil
}

func (r *CountryRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.Country, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, code, name, default_locale, default_currency, default_tz, status,
			data_residency, created_at, updated_at
		FROM global_countries WHERE tenant_id=$1 ORDER BY code ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Country{}
	for rows.Next() {
		c, err := r.scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

type PlaceRepo struct{ DB *sql.DB }

func (r *PlaceRepo) Save(ctx context.Context, p domain.Place) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO global_places (
			id, tenant_id, country_id, parent_id, kind, code, name, active, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (id) DO UPDATE SET
			parent_id=EXCLUDED.parent_id, kind=EXCLUDED.kind, code=EXCLUDED.code,
			name=EXCLUDED.name, active=EXCLUDED.active`,
		p.ID, p.TenantID, p.CountryID, nullUUID(p.ParentID), p.Kind, p.Code, p.Name, p.Active, p.CreatedAt.UTC())
	return err
}

func (r *PlaceRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Place, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, country_id, parent_id, kind, code, name, active, created_at
		FROM global_places WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	var p domain.Place
	var parent uuid.NullUUID
	err := row.Scan(&p.ID, &p.TenantID, &p.CountryID, &parent, &p.Kind, &p.Code, &p.Name, &p.Active, &p.CreatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.Place{}, domain.ErrNotFound
		}
		return domain.Place{}, err
	}
	p.ParentID = scanNullUUID(parent)
	p.CreatedAt = p.CreatedAt.UTC()
	return p, nil
}

func (r *PlaceRepo) ListByCountry(ctx context.Context, tenantID, countryID uuid.UUID) ([]domain.Place, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, country_id, parent_id, kind, code, name, active, created_at
		FROM global_places WHERE tenant_id=$1 AND country_id=$2 ORDER BY code ASC`, tenantID, countryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Place{}
	for rows.Next() {
		var p domain.Place
		var parent uuid.NullUUID
		if err := rows.Scan(&p.ID, &p.TenantID, &p.CountryID, &parent, &p.Kind, &p.Code, &p.Name, &p.Active, &p.CreatedAt); err != nil {
			return nil, err
		}
		p.ParentID = scanNullUUID(parent)
		p.CreatedAt = p.CreatedAt.UTC()
		out = append(out, p)
	}
	return out, rows.Err()
}

type LanguageRepo struct{ DB *sql.DB }

func (r *LanguageRepo) Save(ctx context.Context, l domain.Language) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO global_languages (
			id, tenant_id, code, name, rtl, enabled, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (id) DO UPDATE SET
			code=EXCLUDED.code, name=EXCLUDED.name, rtl=EXCLUDED.rtl, enabled=EXCLUDED.enabled`,
		l.ID, l.TenantID, l.Code, l.Name, l.RTL, l.Enabled, l.CreatedAt.UTC())
	return err
}

func (r *LanguageRepo) GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (domain.Language, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, code, name, rtl, enabled, created_at
		FROM global_languages WHERE tenant_id=$1 AND code=$2`, tenantID, code)
	var l domain.Language
	err := row.Scan(&l.ID, &l.TenantID, &l.Code, &l.Name, &l.RTL, &l.Enabled, &l.CreatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.Language{}, domain.ErrNotFound
		}
		return domain.Language{}, err
	}
	l.CreatedAt = l.CreatedAt.UTC()
	return l, nil
}

func (r *LanguageRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.Language, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, code, name, rtl, enabled, created_at
		FROM global_languages WHERE tenant_id=$1 ORDER BY code ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Language{}
	for rows.Next() {
		var l domain.Language
		if err := rows.Scan(&l.ID, &l.TenantID, &l.Code, &l.Name, &l.RTL, &l.Enabled, &l.CreatedAt); err != nil {
			return nil, err
		}
		l.CreatedAt = l.CreatedAt.UTC()
		out = append(out, l)
	}
	return out, rows.Err()
}

type LocaleRepo struct{ DB *sql.DB }

func (r *LocaleRepo) Save(ctx context.Context, l domain.LocaleProfile) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO global_locales (
			id, tenant_id, locale, language_code, country_code, date_format, time_format,
			number_format, currency_format, first_day_of_week, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (id) DO UPDATE SET
			locale=EXCLUDED.locale, language_code=EXCLUDED.language_code, country_code=EXCLUDED.country_code,
			date_format=EXCLUDED.date_format, time_format=EXCLUDED.time_format,
			number_format=EXCLUDED.number_format, currency_format=EXCLUDED.currency_format,
			first_day_of_week=EXCLUDED.first_day_of_week`,
		l.ID, l.TenantID, l.Locale, l.LanguageCode, l.CountryCode, l.DateFormat, l.TimeFormat,
		l.NumberFormat, l.CurrencyFormat, l.FirstDayOfWeek, l.CreatedAt.UTC())
	return err
}

func (r *LocaleRepo) Get(ctx context.Context, tenantID uuid.UUID, locale string) (domain.LocaleProfile, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, locale, language_code, country_code, date_format, time_format,
			number_format, currency_format, first_day_of_week, created_at
		FROM global_locales WHERE tenant_id=$1 AND locale=$2`, tenantID, locale)
	var l domain.LocaleProfile
	err := row.Scan(&l.ID, &l.TenantID, &l.Locale, &l.LanguageCode, &l.CountryCode, &l.DateFormat, &l.TimeFormat,
		&l.NumberFormat, &l.CurrencyFormat, &l.FirstDayOfWeek, &l.CreatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.LocaleProfile{}, domain.ErrNotFound
		}
		return domain.LocaleProfile{}, err
	}
	l.CreatedAt = l.CreatedAt.UTC()
	return l, nil
}

type TranslationRepo struct{ DB *sql.DB }

func (r *TranslationRepo) Save(ctx context.Context, t domain.Translation) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO global_translations (
			id, tenant_id, namespace, key, locale, value, version, context, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (id) DO UPDATE SET
			namespace=EXCLUDED.namespace, key=EXCLUDED.key, locale=EXCLUDED.locale,
			value=EXCLUDED.value, version=EXCLUDED.version, context=EXCLUDED.context,
			updated_at=EXCLUDED.updated_at`,
		t.ID, t.TenantID, t.Namespace, t.Key, t.Locale, t.Value, t.Version, t.Context, t.UpdatedAt.UTC())
	return err
}

func (r *TranslationRepo) Get(ctx context.Context, tenantID uuid.UUID, ns, key, locale string) (domain.Translation, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, namespace, key, locale, value, version, context, updated_at
		FROM global_translations WHERE tenant_id=$1 AND namespace=$2 AND key=$3 AND locale=$4`,
		tenantID, ns, key, locale)
	var t domain.Translation
	err := row.Scan(&t.ID, &t.TenantID, &t.Namespace, &t.Key, &t.Locale, &t.Value, &t.Version, &t.Context, &t.UpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.Translation{}, domain.ErrNotFound
		}
		return domain.Translation{}, err
	}
	t.UpdatedAt = t.UpdatedAt.UTC()
	return t, nil
}

func (r *TranslationRepo) ListNamespace(ctx context.Context, tenantID uuid.UUID, ns, locale string) ([]domain.Translation, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, namespace, key, locale, value, version, context, updated_at
		FROM global_translations WHERE tenant_id=$1 AND namespace=$2 AND locale=$3 ORDER BY key ASC`,
		tenantID, ns, locale)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Translation{}
	for rows.Next() {
		var t domain.Translation
		if err := rows.Scan(&t.ID, &t.TenantID, &t.Namespace, &t.Key, &t.Locale, &t.Value, &t.Version, &t.Context, &t.UpdatedAt); err != nil {
			return nil, err
		}
		t.UpdatedAt = t.UpdatedAt.UTC()
		out = append(out, t)
	}
	return out, rows.Err()
}

type CurrencyRepo struct{ DB *sql.DB }

func (r *CurrencyRepo) Save(ctx context.Context, c domain.Currency) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO global_currencies (
			id, tenant_id, code, name, minor_units, symbol, enabled, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (id) DO UPDATE SET
			code=EXCLUDED.code, name=EXCLUDED.name, minor_units=EXCLUDED.minor_units,
			symbol=EXCLUDED.symbol, enabled=EXCLUDED.enabled`,
		c.ID, c.TenantID, c.Code, c.Name, c.MinorUnits, c.Symbol, c.Enabled, c.CreatedAt.UTC())
	return err
}

func (r *CurrencyRepo) GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (domain.Currency, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, code, name, minor_units, symbol, enabled, created_at
		FROM global_currencies WHERE tenant_id=$1 AND code=$2`, tenantID, code)
	var c domain.Currency
	err := row.Scan(&c.ID, &c.TenantID, &c.Code, &c.Name, &c.MinorUnits, &c.Symbol, &c.Enabled, &c.CreatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.Currency{}, domain.ErrNotFound
		}
		return domain.Currency{}, err
	}
	c.CreatedAt = c.CreatedAt.UTC()
	return c, nil
}

func (r *CurrencyRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.Currency, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, code, name, minor_units, symbol, enabled, created_at
		FROM global_currencies WHERE tenant_id=$1 ORDER BY code ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Currency{}
	for rows.Next() {
		var c domain.Currency
		if err := rows.Scan(&c.ID, &c.TenantID, &c.Code, &c.Name, &c.MinorUnits, &c.Symbol, &c.Enabled, &c.CreatedAt); err != nil {
			return nil, err
		}
		c.CreatedAt = c.CreatedAt.UTC()
		out = append(out, c)
	}
	return out, rows.Err()
}

type RateRepo struct{ DB *sql.DB }

func (r *RateRepo) Save(ctx context.Context, rate domain.ExchangeRate) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO global_exchange_rates (
			id, tenant_id, base_currency, quote_currency, rate, as_of, source, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (id) DO UPDATE SET
			rate=EXCLUDED.rate, as_of=EXCLUDED.as_of, source=EXCLUDED.source`,
		rate.ID, rate.TenantID, rate.BaseCurrency, rate.QuoteCurrency, rate.Rate,
		rate.AsOf.UTC(), rate.Source, rate.CreatedAt.UTC())
	return err
}

func (r *RateRepo) Latest(ctx context.Context, tenantID uuid.UUID, base, quote string) (domain.ExchangeRate, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, base_currency, quote_currency, rate, as_of, source, created_at
		FROM global_exchange_rates
		WHERE tenant_id=$1 AND base_currency=$2 AND quote_currency=$3
		ORDER BY as_of DESC LIMIT 1`, tenantID, base, quote)
	var rate domain.ExchangeRate
	err := row.Scan(&rate.ID, &rate.TenantID, &rate.BaseCurrency, &rate.QuoteCurrency, &rate.Rate,
		&rate.AsOf, &rate.Source, &rate.CreatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.ExchangeRate{}, domain.ErrNotFound
		}
		return domain.ExchangeRate{}, err
	}
	rate.AsOf = rate.AsOf.UTC()
	rate.CreatedAt = rate.CreatedAt.UTC()
	return rate, nil
}

func (r *RateRepo) History(ctx context.Context, tenantID uuid.UUID, base, quote string, limit int) ([]domain.ExchangeRate, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, base_currency, quote_currency, rate, as_of, source, created_at
		FROM global_exchange_rates
		WHERE tenant_id=$1 AND base_currency=$2 AND quote_currency=$3
		ORDER BY as_of DESC LIMIT $4`, tenantID, base, quote, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ExchangeRate{}
	for rows.Next() {
		var rate domain.ExchangeRate
		if err := rows.Scan(&rate.ID, &rate.TenantID, &rate.BaseCurrency, &rate.QuoteCurrency, &rate.Rate,
			&rate.AsOf, &rate.Source, &rate.CreatedAt); err != nil {
			return nil, err
		}
		rate.AsOf = rate.AsOf.UTC()
		rate.CreatedAt = rate.CreatedAt.UTC()
		out = append(out, rate)
	}
	return out, rows.Err()
}

var (
	_ ports.CountryRepo     = (*CountryRepo)(nil)
	_ ports.PlaceRepo       = (*PlaceRepo)(nil)
	_ ports.LanguageRepo    = (*LanguageRepo)(nil)
	_ ports.LocaleRepo      = (*LocaleRepo)(nil)
	_ ports.TranslationRepo = (*TranslationRepo)(nil)
	_ ports.CurrencyRepo    = (*CurrencyRepo)(nil)
	_ ports.RateRepo        = (*RateRepo)(nil)
)
