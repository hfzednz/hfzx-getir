package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/superapp-service/internal/app/ports"
	"github.com/nexora/superapp-service/internal/domain"
)

type ModuleRepo struct{ DB *sql.DB }

func (r *ModuleRepo) Save(ctx context.Context, m domain.Module) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO sa_modules (
			id, tenant_id, key, name, kind, category, description, publisher_id, status,
			latest_version, icon_uri, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		ON CONFLICT (tenant_id, key) DO UPDATE SET
			id=EXCLUDED.id, name=EXCLUDED.name, kind=EXCLUDED.kind, category=EXCLUDED.category,
			description=EXCLUDED.description, publisher_id=EXCLUDED.publisher_id, status=EXCLUDED.status,
			latest_version=EXCLUDED.latest_version, icon_uri=EXCLUDED.icon_uri, updated_at=EXCLUDED.updated_at`,
		m.ID, m.TenantID, m.Key, m.Name, string(m.Kind), m.Category, m.Description, m.PublisherID, string(m.Status),
		m.LatestVersion, m.IconURI, m.CreatedAt.UTC(), m.UpdatedAt.UTC())
	return err
}
func (r *ModuleRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Module, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, key, name, kind, category, description, publisher_id, status,
			latest_version, icon_uri, created_at, updated_at
		FROM sa_modules WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	return scanModule(row)
}
func (r *ModuleRepo) GetByKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.Module, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, key, name, kind, category, description, publisher_id, status,
			latest_version, icon_uri, created_at, updated_at
		FROM sa_modules WHERE tenant_id=$1 AND key=$2`, tenantID, key)
	return scanModule(row)
}
func (r *ModuleRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.Module, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, key, name, kind, category, description, publisher_id, status,
			latest_version, icon_uri, created_at, updated_at
		FROM sa_modules WHERE tenant_id=$1 ORDER BY key ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Module{}
	for rows.Next() {
		m, err := scanModule(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

type scannable interface{ Scan(dest ...any) error }

func scanModule(row scannable) (domain.Module, error) {
	var m domain.Module
	var kind, status string
	err := row.Scan(&m.ID, &m.TenantID, &m.Key, &m.Name, &kind, &m.Category, &m.Description, &m.PublisherID, &status,
		&m.LatestVersion, &m.IconURI, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.Module{}, domain.ErrNotFound
		}
		return domain.Module{}, err
	}
	m.Kind = domain.ModuleKind(kind)
	m.Status = domain.ModuleStatus(status)
	m.CreatedAt = m.CreatedAt.UTC()
	m.UpdatedAt = m.UpdatedAt.UTC()
	return m, nil
}

type ManifestRepo struct{ DB *sql.DB }

func (r *ManifestRepo) Save(ctx context.Context, m domain.ModuleManifest) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO sa_manifests (
			id, tenant_id, module_id, version, entry_point, min_shell_version, permissions, hooks,
			signature, checksum, bundle_uri, compatible, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		ON CONFLICT (tenant_id, module_id, version) DO UPDATE SET
			id=EXCLUDED.id, entry_point=EXCLUDED.entry_point, min_shell_version=EXCLUDED.min_shell_version,
			permissions=EXCLUDED.permissions, hooks=EXCLUDED.hooks, signature=EXCLUDED.signature,
			checksum=EXCLUDED.checksum, bundle_uri=EXCLUDED.bundle_uri, compatible=EXCLUDED.compatible`,
		m.ID, m.TenantID, m.ModuleID, m.Version, m.EntryPoint, m.MinShellVersion, TextArray(m.Permissions), TextArray(m.Hooks),
		m.Signature, m.Checksum, m.BundleURI, m.Compatible, m.CreatedAt.UTC())
	return err
}
func (r *ManifestRepo) Get(ctx context.Context, tenantID, moduleID uuid.UUID, version string) (domain.ModuleManifest, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, module_id, version, entry_point, min_shell_version, permissions, hooks,
			signature, checksum, bundle_uri, compatible, created_at
		FROM sa_manifests WHERE tenant_id=$1 AND module_id=$2 AND version=$3`, tenantID, moduleID, version)
	return scanManifest(row)
}
func (r *ManifestRepo) Latest(ctx context.Context, tenantID, moduleID uuid.UUID) (domain.ModuleManifest, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, module_id, version, entry_point, min_shell_version, permissions, hooks,
			signature, checksum, bundle_uri, compatible, created_at
		FROM sa_manifests WHERE tenant_id=$1 AND module_id=$2 ORDER BY created_at DESC LIMIT 1`, tenantID, moduleID)
	return scanManifest(row)
}
func scanManifest(row scannable) (domain.ModuleManifest, error) {
	var m domain.ModuleManifest
	var perms, hooks TextArray
	err := row.Scan(&m.ID, &m.TenantID, &m.ModuleID, &m.Version, &m.EntryPoint, &m.MinShellVersion, &perms, &hooks,
		&m.Signature, &m.Checksum, &m.BundleURI, &m.Compatible, &m.CreatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.ModuleManifest{}, domain.ErrNotFound
		}
		return domain.ModuleManifest{}, err
	}
	m.Permissions = []string(perms)
	m.Hooks = []string(hooks)
	m.Dependencies = map[string]string{}
	m.CreatedAt = m.CreatedAt.UTC()
	return m, nil
}

type InstallRepo struct{ DB *sql.DB }

func (r *InstallRepo) Save(ctx context.Context, i domain.Install) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO sa_installs (
			id, tenant_id, subject_id, module_id, version, status, previous_version, installed_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (tenant_id, subject_id, module_id) DO UPDATE SET
			id=EXCLUDED.id, version=EXCLUDED.version, status=EXCLUDED.status,
			previous_version=EXCLUDED.previous_version, updated_at=EXCLUDED.updated_at`,
		i.ID, i.TenantID, i.SubjectID, i.ModuleID, i.Version, string(i.Status), i.PreviousVersion, i.InstalledAt.UTC(), i.UpdatedAt.UTC())
	return err
}
func (r *InstallRepo) Get(ctx context.Context, tenantID uuid.UUID, subjectID string, moduleID uuid.UUID) (domain.Install, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, subject_id, module_id, version, status, previous_version, installed_at, updated_at
		FROM sa_installs WHERE tenant_id=$1 AND subject_id=$2 AND module_id=$3`, tenantID, subjectID, moduleID)
	var i domain.Install
	var status string
	err := row.Scan(&i.ID, &i.TenantID, &i.SubjectID, &i.ModuleID, &i.Version, &status, &i.PreviousVersion, &i.InstalledAt, &i.UpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.Install{}, domain.ErrNotFound
		}
		return domain.Install{}, err
	}
	i.Status = domain.InstallStatus(status)
	i.InstalledAt = i.InstalledAt.UTC()
	i.UpdatedAt = i.UpdatedAt.UTC()
	return i, nil
}
func (r *InstallRepo) ListBySubject(ctx context.Context, tenantID uuid.UUID, subjectID string) ([]domain.Install, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, subject_id, module_id, version, status, previous_version, installed_at, updated_at
		FROM sa_installs WHERE tenant_id=$1 AND subject_id=$2 ORDER BY updated_at DESC`, tenantID, subjectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Install{}
	for rows.Next() {
		var i domain.Install
		var status string
		if err := rows.Scan(&i.ID, &i.TenantID, &i.SubjectID, &i.ModuleID, &i.Version, &status, &i.PreviousVersion, &i.InstalledAt, &i.UpdatedAt); err != nil {
			return nil, err
		}
		i.Status = domain.InstallStatus(status)
		i.InstalledAt = i.InstalledAt.UTC()
		i.UpdatedAt = i.UpdatedAt.UTC()
		out = append(out, i)
	}
	return out, rows.Err()
}

type PermissionRepo struct{ DB *sql.DB }

func (r *PermissionRepo) Save(ctx context.Context, g domain.PermissionGrant) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO sa_permission_grants (id, tenant_id, subject_id, module_id, permission, granted, granted_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (tenant_id, subject_id, module_id, permission) DO UPDATE SET
			id=EXCLUDED.id, granted=EXCLUDED.granted, granted_at=EXCLUDED.granted_at`,
		g.ID, g.TenantID, g.SubjectID, g.ModuleID, g.Permission, g.Granted, g.GrantedAt.UTC())
	return err
}
func (r *PermissionRepo) List(ctx context.Context, tenantID uuid.UUID, subjectID string, moduleID uuid.UUID) ([]domain.PermissionGrant, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, subject_id, module_id, permission, granted, granted_at
		FROM sa_permission_grants WHERE tenant_id=$1 AND subject_id=$2 AND module_id=$3`, tenantID, subjectID, moduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.PermissionGrant{}
	for rows.Next() {
		var g domain.PermissionGrant
		if err := rows.Scan(&g.ID, &g.TenantID, &g.SubjectID, &g.ModuleID, &g.Permission, &g.Granted, &g.GrantedAt); err != nil {
			return nil, err
		}
		g.GrantedAt = g.GrantedAt.UTC()
		out = append(out, g)
	}
	return out, rows.Err()
}
func (r *PermissionRepo) Has(ctx context.Context, tenantID uuid.UUID, subjectID string, moduleID uuid.UUID, perm string) (bool, error) {
	var granted bool
	err := r.DB.QueryRowContext(ctx, `
		SELECT granted FROM sa_permission_grants
		WHERE tenant_id=$1 AND subject_id=$2 AND module_id=$3 AND permission=$4`, tenantID, subjectID, moduleID, perm).Scan(&granted)
	if err != nil {
		if isNoRows(err) {
			return false, nil
		}
		return false, err
	}
	return granted, nil
}

type ListingRepo struct{ DB *sql.DB }

func (r *ListingRepo) Save(ctx context.Context, l domain.StoreListing) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO sa_store_listings (
			id, tenant_id, module_id, featured, price_minor, currency, subscription,
			rating_avg, rating_count, installs, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (tenant_id, module_id) DO UPDATE SET
			id=EXCLUDED.id, featured=EXCLUDED.featured, price_minor=EXCLUDED.price_minor,
			currency=EXCLUDED.currency, subscription=EXCLUDED.subscription, rating_avg=EXCLUDED.rating_avg,
			rating_count=EXCLUDED.rating_count, installs=EXCLUDED.installs, updated_at=EXCLUDED.updated_at`,
		l.ID, l.TenantID, l.ModuleID, l.Featured, l.PriceMinor, l.Currency, l.Subscription,
		l.RatingAvg, l.RatingCount, l.Installs, l.UpdatedAt.UTC())
	return err
}
func (r *ListingRepo) GetByModule(ctx context.Context, tenantID, moduleID uuid.UUID) (domain.StoreListing, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, module_id, featured, price_minor, currency, subscription,
			rating_avg, rating_count, installs, updated_at
		FROM sa_store_listings WHERE tenant_id=$1 AND module_id=$2`, tenantID, moduleID)
	var l domain.StoreListing
	err := row.Scan(&l.ID, &l.TenantID, &l.ModuleID, &l.Featured, &l.PriceMinor, &l.Currency, &l.Subscription,
		&l.RatingAvg, &l.RatingCount, &l.Installs, &l.UpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.StoreListing{}, domain.ErrNotFound
		}
		return domain.StoreListing{}, err
	}
	l.UpdatedAt = l.UpdatedAt.UTC()
	return l, nil
}
func (r *ListingRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.StoreListing, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, module_id, featured, price_minor, currency, subscription,
			rating_avg, rating_count, installs, updated_at
		FROM sa_store_listings WHERE tenant_id=$1 ORDER BY featured DESC, installs DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.StoreListing{}
	for rows.Next() {
		var l domain.StoreListing
		if err := rows.Scan(&l.ID, &l.TenantID, &l.ModuleID, &l.Featured, &l.PriceMinor, &l.Currency, &l.Subscription,
			&l.RatingAvg, &l.RatingCount, &l.Installs, &l.UpdatedAt); err != nil {
			return nil, err
		}
		l.UpdatedAt = l.UpdatedAt.UTC()
		out = append(out, l)
	}
	return out, rows.Err()
}

type RatingRepo struct{ DB *sql.DB }

func (r *RatingRepo) Save(ctx context.Context, rating domain.StoreRating) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO sa_store_ratings (id, tenant_id, module_id, subject_id, score, comment, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		rating.ID, rating.TenantID, rating.ModuleID, rating.SubjectID, rating.Score, rating.Comment, rating.CreatedAt.UTC())
	return err
}
func (r *RatingRepo) ListByModule(ctx context.Context, tenantID, moduleID uuid.UUID) ([]domain.StoreRating, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, module_id, subject_id, score, comment, created_at
		FROM sa_store_ratings WHERE tenant_id=$1 AND module_id=$2 ORDER BY created_at DESC`, tenantID, moduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.StoreRating{}
	for rows.Next() {
		var rating domain.StoreRating
		if err := rows.Scan(&rating.ID, &rating.TenantID, &rating.ModuleID, &rating.SubjectID, &rating.Score, &rating.Comment, &rating.CreatedAt); err != nil {
			return nil, err
		}
		rating.CreatedAt = rating.CreatedAt.UTC()
		out = append(out, rating)
	}
	return out, rows.Err()
}

type WidgetRepo struct{ DB *sql.DB }

func (r *WidgetRepo) Save(ctx context.Context, w domain.WidgetPlacement) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO sa_widgets (id, tenant_id, subject_id, module_id, slot, sort_order, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (id) DO UPDATE SET slot=EXCLUDED.slot, sort_order=EXCLUDED.sort_order`,
		w.ID, w.TenantID, w.SubjectID, w.ModuleID, w.Slot, w.Order, w.CreatedAt.UTC())
	return err
}
func (r *WidgetRepo) ListBySubject(ctx context.Context, tenantID uuid.UUID, subjectID string) ([]domain.WidgetPlacement, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, subject_id, module_id, slot, sort_order, created_at
		FROM sa_widgets WHERE tenant_id=$1 AND subject_id=$2 ORDER BY sort_order ASC`, tenantID, subjectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.WidgetPlacement{}
	for rows.Next() {
		var w domain.WidgetPlacement
		if err := rows.Scan(&w.ID, &w.TenantID, &w.SubjectID, &w.ModuleID, &w.Slot, &w.Order, &w.CreatedAt); err != nil {
			return nil, err
		}
		w.CreatedAt = w.CreatedAt.UTC()
		out = append(out, w)
	}
	return out, rows.Err()
}

type MonetizationRepo struct{ DB *sql.DB }

func (r *MonetizationRepo) Save(ctx context.Context, rule domain.MonetizationRule) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO sa_monetization (id, tenant_id, module_id, commission_bps, partner_share_bps, active, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (tenant_id, module_id) DO UPDATE SET
			id=EXCLUDED.id, commission_bps=EXCLUDED.commission_bps, partner_share_bps=EXCLUDED.partner_share_bps,
			active=EXCLUDED.active, updated_at=EXCLUDED.updated_at`,
		rule.ID, rule.TenantID, rule.ModuleID, rule.CommissionBps, rule.PartnerShareBps, rule.Active, rule.UpdatedAt.UTC())
	return err
}
func (r *MonetizationRepo) GetByModule(ctx context.Context, tenantID, moduleID uuid.UUID) (domain.MonetizationRule, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, module_id, commission_bps, partner_share_bps, active, updated_at
		FROM sa_monetization WHERE tenant_id=$1 AND module_id=$2`, tenantID, moduleID)
	var rule domain.MonetizationRule
	err := row.Scan(&rule.ID, &rule.TenantID, &rule.ModuleID, &rule.CommissionBps, &rule.PartnerShareBps, &rule.Active, &rule.UpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.MonetizationRule{}, domain.ErrNotFound
		}
		return domain.MonetizationRule{}, err
	}
	rule.UpdatedAt = rule.UpdatedAt.UTC()
	return rule, nil
}

type LaunchRepo struct{ DB *sql.DB }

func (r *LaunchRepo) Save(ctx context.Context, e domain.LaunchEvent) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO sa_launches (id, tenant_id, subject_id, module_id, kind, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		e.ID, e.TenantID, e.SubjectID, e.ModuleID, string(e.Kind), e.CreatedAt.UTC())
	return err
}

var (
	_ ports.ModuleRepo      = (*ModuleRepo)(nil)
	_ ports.ManifestRepo    = (*ManifestRepo)(nil)
	_ ports.InstallRepo     = (*InstallRepo)(nil)
	_ ports.PermissionRepo  = (*PermissionRepo)(nil)
	_ ports.ListingRepo     = (*ListingRepo)(nil)
	_ ports.RatingRepo      = (*RatingRepo)(nil)
	_ ports.WidgetRepo      = (*WidgetRepo)(nil)
	_ ports.MonetizationRepo = (*MonetizationRepo)(nil)
	_ ports.LaunchRepo      = (*LaunchRepo)(nil)
)
