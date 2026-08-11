package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/nexora/catalog-service/internal/app/ports"
	"github.com/nexora/catalog-service/internal/domain"
)

type ImportJobRepo struct{ DB *sql.DB }

func (r *ImportJobRepo) Create(ctx context.Context, j domain.ImportJob) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO import_jobs (
			id, tenant_id, kind, status, source_format, source_uri, result_uri, total_rows, processed_rows,
			success_rows, error_rows, errors, options, created_by, started_at, finished_at, created_at, updated_at
		) VALUES ($1,$2,$3::import_job_kind,$4::import_job_status,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)`,
		j.ID, j.TenantID, string(j.Kind), string(j.Status), j.SourceFormat, j.SourceURI, j.ResultURI, j.TotalRows, j.ProcessedRows,
		j.SuccessRows, j.ErrorRows, JSONArray(j.Errors), JSONMap(j.Options), nullUUID(j.CreatedBy),
		nullTime(j.StartedAt), nullTime(j.FinishedAt), j.CreatedAt, j.UpdatedAt)
	return err
}

func (r *ImportJobRepo) Update(ctx context.Context, j domain.ImportJob) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE import_jobs SET kind=$2::import_job_kind, status=$3::import_job_status, source_format=$4, source_uri=$5,
			result_uri=$6, total_rows=$7, processed_rows=$8, success_rows=$9, error_rows=$10, errors=$11, options=$12,
			created_by=$13, started_at=$14, finished_at=$15, updated_at=$16
		WHERE id=$1 AND tenant_id=$17`,
		j.ID, string(j.Kind), string(j.Status), j.SourceFormat, j.SourceURI,
		j.ResultURI, j.TotalRows, j.ProcessedRows, j.SuccessRows, j.ErrorRows, JSONArray(j.Errors), JSONMap(j.Options),
		nullUUID(j.CreatedBy), nullTime(j.StartedAt), nullTime(j.FinishedAt), j.UpdatedAt, j.TenantID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ImportJobRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.ImportJob, error) {
	j, err := scanImportJob(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, kind::text, status::text, source_format, source_uri, result_uri, total_rows, processed_rows,
			success_rows, error_rows, errors, options, created_by, started_at, finished_at, created_at, updated_at
		FROM import_jobs WHERE id=$1 AND tenant_id=$2`, id, tenantID))
	if err != nil {
		return domain.ImportJob{}, mapNotFound(err)
	}
	return j, nil
}

func (r *ImportJobRepo) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.ImportJob, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, kind::text, status::text, source_format, source_uri, result_uri, total_rows, processed_rows,
			success_rows, error_rows, errors, options, created_by, started_at, finished_at, created_at, updated_at
		FROM import_jobs WHERE tenant_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ImportJob{}
	for rows.Next() {
		j, err := scanImportJob(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, rows.Err()
}

type importJobScanner interface {
	Scan(dest ...any) error
}

func scanImportJob(s importJobScanner) (domain.ImportJob, error) {
	var j domain.ImportJob
	var kind, status string
	var errs JSONArray
	var opts JSONMap
	var createdBy uuid.NullUUID
	var started, finished sql.NullTime
	err := s.Scan(&j.ID, &j.TenantID, &kind, &status, &j.SourceFormat, &j.SourceURI, &j.ResultURI, &j.TotalRows, &j.ProcessedRows,
		&j.SuccessRows, &j.ErrorRows, &errs, &opts, &createdBy, &started, &finished, &j.CreatedAt, &j.UpdatedAt)
	if err != nil {
		return domain.ImportJob{}, err
	}
	j.Kind = domain.ImportJobKind(kind)
	j.Status = domain.ImportJobStatus(status)
	j.Errors = []map[string]any(errs)
	j.Options = map[string]any(opts)
	j.CreatedBy = scanNullUUID(createdBy)
	j.StartedAt = scanNullTime(started)
	j.FinishedAt = scanNullTime(finished)
	return j, nil
}

var _ ports.ImportJobRepository = (*ImportJobRepo)(nil)

type ComplianceRepo struct{ DB *sql.DB }

func (r *ComplianceRepo) Upsert(ctx context.Context, c domain.ProductCompliance) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO product_compliance (
			id, product_id, tenant_id, age_restriction, is_hazardous, hazard_class, is_pharmacy, requires_prescription,
			is_food, is_organic, is_halal, is_vegan, is_gluten_free, restricted_countries, allowed_countries,
			certificates, metadata, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)
		ON CONFLICT (product_id) DO UPDATE SET
			age_restriction=EXCLUDED.age_restriction, is_hazardous=EXCLUDED.is_hazardous, hazard_class=EXCLUDED.hazard_class,
			is_pharmacy=EXCLUDED.is_pharmacy, requires_prescription=EXCLUDED.requires_prescription, is_food=EXCLUDED.is_food,
			is_organic=EXCLUDED.is_organic, is_halal=EXCLUDED.is_halal, is_vegan=EXCLUDED.is_vegan,
			is_gluten_free=EXCLUDED.is_gluten_free, restricted_countries=EXCLUDED.restricted_countries,
			allowed_countries=EXCLUDED.allowed_countries, certificates=EXCLUDED.certificates, metadata=EXCLUDED.metadata,
			updated_at=EXCLUDED.updated_at, id=product_compliance.id`,
		c.ID, c.ProductID, c.TenantID, c.AgeRestriction, c.IsHazardous, c.HazardClass, c.IsPharmacy, c.RequiresPrescription,
		c.IsFood, c.IsOrganic, c.IsHalal, c.IsVegan, c.IsGlutenFree, textArray(c.RestrictedCountries), textArray(c.AllowedCountries),
		JSONArray(c.Certificates), JSONMap(c.Metadata), c.CreatedAt, c.UpdatedAt)
	return err
}

func (r *ComplianceRepo) GetByProduct(ctx context.Context, tenantID, productID uuid.UUID) (domain.ProductCompliance, error) {
	var c domain.ProductCompliance
	var certs JSONArray
	var meta JSONMap
	var restricted, allowed []string
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, product_id, tenant_id, age_restriction, is_hazardous, hazard_class, is_pharmacy, requires_prescription,
			is_food, is_organic, is_halal, is_vegan, is_gluten_free, restricted_countries, allowed_countries,
			certificates, metadata, created_at, updated_at
		FROM product_compliance WHERE product_id=$1 AND tenant_id=$2`, productID, tenantID).
		Scan(&c.ID, &c.ProductID, &c.TenantID, &c.AgeRestriction, &c.IsHazardous, &c.HazardClass, &c.IsPharmacy, &c.RequiresPrescription,
			&c.IsFood, &c.IsOrganic, &c.IsHalal, &c.IsVegan, &c.IsGlutenFree, pq.Array(&restricted), pq.Array(&allowed),
			&certs, &meta, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return domain.ProductCompliance{}, mapNotFound(err)
	}
	c.RestrictedCountries = restricted
	c.AllowedCountries = allowed
	c.Certificates = []map[string]any(certs)
	c.Metadata = map[string]any(meta)
	return c, nil
}

var _ ports.ComplianceRepository = (*ComplianceRepo)(nil)
