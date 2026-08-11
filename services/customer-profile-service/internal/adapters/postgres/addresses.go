package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/customer-profile-service/internal/app/ports"
	"github.com/nexora/customer-profile-service/internal/domain"
)

// AddressRepo persists addresses.
type AddressRepo struct{ DB *sql.DB }

var _ ports.AddressRepository = (*AddressRepo)(nil)

func (r *AddressRepo) Create(ctx context.Context, a domain.Address) error {
	label := a.Label
	if label == "" {
		label = domain.AddressLabelHome
	}
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO addresses (
			id, profile_id, tenant_id, label, custom_label, line1, building, apartment, entrance,
			floor, door, notes, lat, lng, city_id, zone_validated_at, is_default,
			created_at, updated_at, deleted_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,
			$10,$11,$12,$13,$14,$15,$16,$17,
			$18,$19,$20
		)`,
		a.ID, a.ProfileID, a.TenantID, string(label), a.CustomLabel, a.Line1, a.Building, a.Apartment, a.Entrance,
		a.Floor, a.Door, a.Notes, a.Lat, a.Lng, nullUUID(a.CityID), nullTime(a.ZoneValidatedAt), a.IsDefault,
		a.CreatedAt, a.UpdatedAt, nullTime(a.DeletedAt),
	)
	return mapUniqueViolation(err)
}

func (r *AddressRepo) Update(ctx context.Context, a domain.Address) error {
	label := a.Label
	if label == "" {
		label = domain.AddressLabelHome
	}
	res, err := r.DB.ExecContext(ctx, `
		UPDATE addresses SET
			label=$1, custom_label=$2, line1=$3, building=$4, apartment=$5, entrance=$6,
			floor=$7, door=$8, notes=$9, lat=$10, lng=$11, city_id=$12, zone_validated_at=$13,
			is_default=$14, updated_at=$15, deleted_at=$16
		WHERE id=$17`,
		string(label), a.CustomLabel, a.Line1, a.Building, a.Apartment, a.Entrance,
		a.Floor, a.Door, a.Notes, a.Lat, a.Lng, nullUUID(a.CityID), nullTime(a.ZoneValidatedAt),
		a.IsDefault, a.UpdatedAt, nullTime(a.DeletedAt), a.ID,
	)
	return rowsAffectedOrNotFound(res, err)
}

func (r *AddressRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.Address, error) {
	return r.scanAddress(r.DB.QueryRowContext(ctx, addressSelect+` WHERE id=$1`, id))
}

func (r *AddressRepo) ListByProfile(ctx context.Context, profileID uuid.UUID) ([]domain.Address, error) {
	rows, err := r.DB.QueryContext(ctx, addressSelect+`
		WHERE profile_id=$1 AND deleted_at IS NULL
		ORDER BY is_default DESC, created_at ASC`, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.Address, 0)
	for rows.Next() {
		a, err := scanAddressRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *AddressRepo) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.DB.ExecContext(ctx, `DELETE FROM addresses WHERE id=$1`, id)
	return rowsAffectedOrNotFound(res, err)
}

func (r *AddressRepo) ClearDefault(ctx context.Context, profileID uuid.UUID) error {
	_, err := r.DB.ExecContext(ctx, `
		UPDATE addresses SET is_default=FALSE, updated_at=now()
		WHERE profile_id=$1 AND is_default=TRUE AND deleted_at IS NULL`, profileID)
	return err
}

const addressSelect = `
	SELECT id, profile_id, tenant_id, label, custom_label, line1, building, apartment, entrance,
		floor, door, notes, lat, lng, city_id, zone_validated_at, is_default,
		created_at, updated_at, deleted_at
	FROM addresses`

func (r *AddressRepo) scanAddress(row scannable) (domain.Address, error) {
	a, err := scanAddressRow(row)
	if err != nil {
		return domain.Address{}, mapNotFound(err)
	}
	return a, nil
}

func scanAddressRow(row scannable) (domain.Address, error) {
	var a domain.Address
	var label string
	var cityID uuid.NullUUID
	var zoneValidated, deleted sql.NullTime
	err := row.Scan(
		&a.ID, &a.ProfileID, &a.TenantID, &label, &a.CustomLabel, &a.Line1, &a.Building, &a.Apartment, &a.Entrance,
		&a.Floor, &a.Door, &a.Notes, &a.Lat, &a.Lng, &cityID, &zoneValidated, &a.IsDefault,
		&a.CreatedAt, &a.UpdatedAt, &deleted,
	)
	if err != nil {
		return domain.Address{}, err
	}
	a.Label = domain.AddressLabel(label)
	a.CityID = scanNullUUID(cityID)
	a.ZoneValidatedAt = scanNullTime(zoneValidated)
	a.DeletedAt = scanNullTime(deleted)
	a.CreatedAt = a.CreatedAt.UTC()
	a.UpdatedAt = a.UpdatedAt.UTC()
	return a, nil
}
