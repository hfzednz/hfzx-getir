package postgres

import (
	"context"
	"database/sql"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/loyalty-service/internal/app/ports"
	"github.com/nexora/loyalty-service/internal/domain"
)

// ReferralRepo persists referral codes and events.
type ReferralRepo struct{ DB *sql.DB }

func (r *ReferralRepo) CreateCode(ctx context.Context, c domain.ReferralCode) error {
	c.Code = strings.ToUpper(c.Code)
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO referral_codes (id, tenant_id, account_id, principal_id, code, active, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		c.ID, c.TenantID, c.AccountID, c.PrincipalID, c.Code, c.Active, c.CreatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *ReferralRepo) GetCode(ctx context.Context, tenantID uuid.UUID, code string) (domain.ReferralCode, error) {
	return r.scanCode(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, account_id, principal_id, code, active, created_at
		FROM referral_codes WHERE tenant_id=$1 AND code=$2`, tenantID, strings.ToUpper(code)))
}

func (r *ReferralRepo) GetCodeByAccount(ctx context.Context, tenantID, accountID uuid.UUID) (domain.ReferralCode, error) {
	return r.scanCode(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, account_id, principal_id, code, active, created_at
		FROM referral_codes WHERE tenant_id=$1 AND account_id=$2`, tenantID, accountID))
}

func (r *ReferralRepo) CreateEvent(ctx context.Context, e domain.ReferralEvent) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO referral_events
		  (id, tenant_id, code_id, referrer_account, referee_account, referee_principal,
		   status, order_id, points_granted, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		e.ID, e.TenantID, e.CodeID, e.ReferrerAccount, e.RefereeAccount, e.RefereePrincipal,
		string(e.Status), nullUUID(e.OrderID), e.PointsGranted, e.CreatedAt.UTC(), e.UpdatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *ReferralRepo) GetEventByReferee(ctx context.Context, tenantID, refereeAccount uuid.UUID) (domain.ReferralEvent, error) {
	return r.scanEvent(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, code_id, referrer_account, referee_account, referee_principal,
		       status, order_id, points_granted, created_at, updated_at
		FROM referral_events WHERE tenant_id=$1 AND referee_account=$2`, tenantID, refereeAccount))
}

func (r *ReferralRepo) UpdateEvent(ctx context.Context, e domain.ReferralEvent) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE referral_events
		SET status=$3, order_id=$4, points_granted=$5, updated_at=$6
		WHERE id=$1 AND tenant_id=$2`,
		e.ID, e.TenantID, string(e.Status), nullUUID(e.OrderID), e.PointsGranted, e.UpdatedAt.UTC())
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ReferralRepo) CountCompletedByReferrer(ctx context.Context, tenantID, referrerAccount uuid.UUID) (int64, error) {
	var n int64
	err := r.DB.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM referral_events
		WHERE tenant_id=$1 AND referrer_account=$2 AND status=$3`,
		tenantID, referrerAccount, string(domain.ReferralCompleted)).Scan(&n)
	return n, err
}

func (r *ReferralRepo) scanCode(row *sql.Row) (domain.ReferralCode, error) {
	var c domain.ReferralCode
	err := row.Scan(&c.ID, &c.TenantID, &c.AccountID, &c.PrincipalID, &c.Code, &c.Active, &c.CreatedAt)
	if isNoRows(err) {
		return domain.ReferralCode{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.ReferralCode{}, err
	}
	c.CreatedAt = c.CreatedAt.UTC()
	return c, nil
}

func (r *ReferralRepo) scanEvent(row *sql.Row) (domain.ReferralEvent, error) {
	var e domain.ReferralEvent
	var status string
	var orderID uuid.NullUUID
	err := row.Scan(
		&e.ID, &e.TenantID, &e.CodeID, &e.ReferrerAccount, &e.RefereeAccount, &e.RefereePrincipal,
		&status, &orderID, &e.PointsGranted, &e.CreatedAt, &e.UpdatedAt)
	if isNoRows(err) {
		return domain.ReferralEvent{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.ReferralEvent{}, err
	}
	e.Status = domain.ReferralStatus(status)
	e.OrderID = scanNullUUID(orderID)
	e.CreatedAt = e.CreatedAt.UTC()
	e.UpdatedAt = e.UpdatedAt.UTC()
	return e, nil
}

var _ ports.ReferralRepo = (*ReferralRepo)(nil)
