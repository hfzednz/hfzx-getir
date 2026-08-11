package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/customer-profile-service/internal/app/ports"
	"github.com/nexora/customer-profile-service/internal/domain"
)

// HouseholdRepo persists households and members.
type HouseholdRepo struct{ DB *sql.DB }

var _ ports.HouseholdRepository = (*HouseholdRepo)(nil)

func (r *HouseholdRepo) Create(ctx context.Context, h domain.Household) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO households (
			id, tenant_id, name, owner_profile_id, share_addresses, share_payments, share_lists,
			share_wallet, share_loyalty, created_at, updated_at, deleted_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		h.ID, h.TenantID, h.Name, h.OwnerProfileID, h.Sharing.Addresses, h.Sharing.Payments, h.Sharing.Lists,
		h.Sharing.Wallet, h.Sharing.Loyalty, h.CreatedAt, h.UpdatedAt, nullTime(h.DeletedAt),
	)
	return mapUniqueViolation(err)
}

func (r *HouseholdRepo) Update(ctx context.Context, h domain.Household) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE households SET
			name=$1, owner_profile_id=$2, share_addresses=$3, share_payments=$4, share_lists=$5,
			share_wallet=$6, share_loyalty=$7, updated_at=$8, deleted_at=$9
		WHERE id=$10`,
		h.Name, h.OwnerProfileID, h.Sharing.Addresses, h.Sharing.Payments, h.Sharing.Lists,
		h.Sharing.Wallet, h.Sharing.Loyalty, h.UpdatedAt, nullTime(h.DeletedAt), h.ID,
	)
	return rowsAffectedOrNotFound(res, err)
}

func (r *HouseholdRepo) Get(ctx context.Context, id uuid.UUID) (domain.Household, error) {
	return r.scanHousehold(r.DB.QueryRowContext(ctx, householdSelect+` WHERE id=$1`, id))
}

func (r *HouseholdRepo) GetByOwner(ctx context.Context, ownerProfileID uuid.UUID) (domain.Household, error) {
	return r.scanHousehold(r.DB.QueryRowContext(ctx, householdSelect+`
		WHERE owner_profile_id=$1 AND deleted_at IS NULL
		ORDER BY created_at ASC LIMIT 1`, ownerProfileID))
}

func (r *HouseholdRepo) AddMember(ctx context.Context, m domain.HouseholdMember) error {
	role := m.Role
	if role == "" {
		role = domain.HouseholdRoleAdult
	}
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO household_members (id, household_id, profile_id, role, joined_at, left_at)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		m.ID, m.HouseholdID, m.ProfileID, string(role), m.JoinedAt, nullTime(m.LeftAt),
	)
	return mapUniqueViolation(err)
}

func (r *HouseholdRepo) UpdateMember(ctx context.Context, m domain.HouseholdMember) error {
	role := m.Role
	if role == "" {
		role = domain.HouseholdRoleAdult
	}
	res, err := r.DB.ExecContext(ctx, `
		UPDATE household_members SET role=$1, joined_at=$2, left_at=$3
		WHERE id=$4`, string(role), m.JoinedAt, nullTime(m.LeftAt), m.ID)
	return rowsAffectedOrNotFound(res, err)
}

func (r *HouseholdRepo) ListMembers(ctx context.Context, householdID uuid.UUID) ([]domain.HouseholdMember, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, household_id, profile_id, role, joined_at, left_at
		FROM household_members WHERE household_id=$1 ORDER BY joined_at`, householdID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.HouseholdMember, 0)
	for rows.Next() {
		m, err := scanMemberRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *HouseholdRepo) GetMember(ctx context.Context, householdID, profileID uuid.UUID) (domain.HouseholdMember, error) {
	m, err := scanMemberRow(r.DB.QueryRowContext(ctx, `
		SELECT id, household_id, profile_id, role, joined_at, left_at
		FROM household_members WHERE household_id=$1 AND profile_id=$2`, householdID, profileID))
	if err != nil {
		return domain.HouseholdMember{}, mapNotFound(err)
	}
	return m, nil
}

const householdSelect = `
	SELECT id, tenant_id, name, owner_profile_id, share_addresses, share_payments, share_lists,
		share_wallet, share_loyalty, created_at, updated_at, deleted_at
	FROM households`

func (r *HouseholdRepo) scanHousehold(row scannable) (domain.Household, error) {
	var h domain.Household
	var deleted sql.NullTime
	err := row.Scan(
		&h.ID, &h.TenantID, &h.Name, &h.OwnerProfileID,
		&h.Sharing.Addresses, &h.Sharing.Payments, &h.Sharing.Lists, &h.Sharing.Wallet, &h.Sharing.Loyalty,
		&h.CreatedAt, &h.UpdatedAt, &deleted,
	)
	if err != nil {
		return domain.Household{}, mapNotFound(err)
	}
	h.DeletedAt = scanNullTime(deleted)
	h.CreatedAt = h.CreatedAt.UTC()
	h.UpdatedAt = h.UpdatedAt.UTC()
	return h, nil
}

func scanMemberRow(row scannable) (domain.HouseholdMember, error) {
	var m domain.HouseholdMember
	var role string
	var left sql.NullTime
	err := row.Scan(&m.ID, &m.HouseholdID, &m.ProfileID, &role, &m.JoinedAt, &left)
	if err != nil {
		return domain.HouseholdMember{}, err
	}
	m.Role = domain.HouseholdMemberRole(role)
	m.LeftAt = scanNullTime(left)
	m.JoinedAt = m.JoinedAt.UTC()
	return m, nil
}
