package postgres

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/customer-profile-service/internal/app/ports"
	"github.com/nexora/customer-profile-service/internal/domain"
)

// ProfileRepo persists customer profiles.
type ProfileRepo struct{ DB *sql.DB }

var _ ports.ProfileRepository = (*ProfileRepo)(nil)

func (r *ProfileRepo) Create(ctx context.Context, p domain.CustomerProfile) error {
	gender := p.Gender
	if gender == "" {
		gender = domain.GenderUnspecified
	}
	status := p.Status
	if status == "" {
		status = domain.ProfileStatusActive
	}
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO customer_profiles (
			id, principal_id, tenant_id, display_name, full_name, nickname, avatar_url, gender,
			birthday, language, country_code, city, timezone, occupation, family_size,
			dietary, accessibility, status, created_at, updated_at, deleted_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,
			$9,$10,$11,$12,$13,$14,$15,
			$16,$17,$18,$19,$20,$21
		)`,
		p.ID, p.PrincipalID, p.TenantID, p.DisplayName, p.FullName, p.Nickname, p.AvatarURL, string(gender),
		nullTime(p.Birthday), p.Language, p.CountryCode, p.City, p.Timezone, p.Occupation, p.FamilySize,
		JSONMap(metaGetMap(p.Dietary)), JSONMap(metaGetMap(p.Accessibility)), string(status),
		p.CreatedAt, p.UpdatedAt, nullTime(p.DeletedAt),
	)
	return mapUniqueViolation(err)
}

func (r *ProfileRepo) Update(ctx context.Context, p domain.CustomerProfile) error {
	gender := p.Gender
	if gender == "" {
		gender = domain.GenderUnspecified
	}
	status := p.Status
	if status == "" {
		status = domain.ProfileStatusActive
	}
	res, err := r.DB.ExecContext(ctx, `
		UPDATE customer_profiles SET
			display_name=$1, full_name=$2, nickname=$3, avatar_url=$4, gender=$5, birthday=$6,
			language=$7, country_code=$8, city=$9, timezone=$10, occupation=$11, family_size=$12,
			dietary=$13, accessibility=$14, status=$15, updated_at=$16, deleted_at=$17
		WHERE id=$18`,
		p.DisplayName, p.FullName, p.Nickname, p.AvatarURL, string(gender), nullTime(p.Birthday),
		p.Language, p.CountryCode, p.City, p.Timezone, p.Occupation, p.FamilySize,
		JSONMap(metaGetMap(p.Dietary)), JSONMap(metaGetMap(p.Accessibility)), string(status),
		p.UpdatedAt, nullTime(p.DeletedAt), p.ID,
	)
	return rowsAffectedOrNotFound(res, err)
}

func (r *ProfileRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.CustomerProfile, error) {
	return r.scanProfile(r.DB.QueryRowContext(ctx, profileSelect+` WHERE id=$1`, id))
}

func (r *ProfileRepo) GetByPrincipalID(ctx context.Context, tenantID, principalID uuid.UUID) (domain.CustomerProfile, error) {
	return r.scanProfile(r.DB.QueryRowContext(ctx, profileSelect+` WHERE tenant_id=$1 AND principal_id=$2`, tenantID, principalID))
}

func (r *ProfileRepo) SoftDelete(ctx context.Context, id uuid.UUID, at time.Time) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE customer_profiles SET status='deleted', deleted_at=$1, updated_at=$1 WHERE id=$2`, at, id)
	return rowsAffectedOrNotFound(res, err)
}

func (r *ProfileRepo) Search(ctx context.Context, tenantID uuid.UUID, query string, limit int) ([]domain.CustomerProfile, error) {
	if limit <= 0 {
		limit = 50
	}
	q := strings.TrimSpace(query)
	rows, err := r.DB.QueryContext(ctx, profileSelect+`
		WHERE tenant_id=$1 AND status='active' AND deleted_at IS NULL
			AND (
				$2 = '' OR
				display_name ILIKE '%' || $2 || '%' OR
				full_name ILIKE '%' || $2 || '%' OR
				nickname ILIKE '%' || $2 || '%' OR
				city ILIKE '%' || $2 || '%' OR
				id::text ILIKE '%' || $2 || '%'
			)
		ORDER BY updated_at DESC
		LIMIT $3`, tenantID, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanProfiles(rows)
}

func (r *ProfileRepo) FindDuplicates(ctx context.Context, tenantID uuid.UUID, displayName, fullName string) ([]domain.CustomerProfile, error) {
	dn := strings.ToLower(strings.TrimSpace(displayName))
	fn := strings.ToLower(strings.TrimSpace(fullName))
	rows, err := r.DB.QueryContext(ctx, profileSelect+`
		WHERE tenant_id=$1 AND status='active' AND deleted_at IS NULL
			AND (
				($2 <> '' AND lower(display_name) = $2) OR
				($3 <> '' AND lower(full_name) = $3)
			)`, tenantID, dn, fn)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanProfiles(rows)
}

const profileSelect = `
	SELECT id, principal_id, tenant_id, display_name, full_name, nickname, avatar_url, gender,
		birthday, language, country_code, city, timezone, occupation, family_size,
		dietary, accessibility, status, created_at, updated_at, deleted_at
	FROM customer_profiles`

func (r *ProfileRepo) scanProfile(row scannable) (domain.CustomerProfile, error) {
	p, err := scanProfileRow(row)
	if err != nil {
		return domain.CustomerProfile{}, mapNotFound(err)
	}
	return p, nil
}

func scanProfiles(rows *sql.Rows) ([]domain.CustomerProfile, error) {
	out := make([]domain.CustomerProfile, 0)
	for rows.Next() {
		p, err := scanProfileRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func scanProfileRow(row scannable) (domain.CustomerProfile, error) {
	var p domain.CustomerProfile
	var gender, status string
	var birthday, deleted sql.NullTime
	var dietary, accessibility JSONMap
	err := row.Scan(
		&p.ID, &p.PrincipalID, &p.TenantID, &p.DisplayName, &p.FullName, &p.Nickname, &p.AvatarURL, &gender,
		&birthday, &p.Language, &p.CountryCode, &p.City, &p.Timezone, &p.Occupation, &p.FamilySize,
		&dietary, &accessibility, &status, &p.CreatedAt, &p.UpdatedAt, &deleted,
	)
	if err != nil {
		return domain.CustomerProfile{}, err
	}
	p.Gender = domain.Gender(gender)
	p.Status = domain.ProfileStatus(status)
	p.Birthday = scanNullTime(birthday)
	p.DeletedAt = scanNullTime(deleted)
	p.Dietary = map[string]any(dietary)
	p.Accessibility = map[string]any(accessibility)
	p.CreatedAt = p.CreatedAt.UTC()
	p.UpdatedAt = p.UpdatedAt.UTC()
	return p, nil
}
