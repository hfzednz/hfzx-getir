package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/payment-service/internal/app/ports"
	"github.com/nexora/payment-service/internal/domain"
)

// IntentRepo persists payment intents and related aggregates.
type IntentRepo struct{ DB *sql.DB }

var _ ports.IntentRepo = (*IntentRepo)(nil)

const intentColumns = `
	id, tenant_id, principal_id, order_id, status, amount_minor, captured_minor, refunded_minor,
	currency, method_type, payment_method_id, provider, provider_intent_ref, idempotency_key,
	fraud_score, fraud_decision, failure_reason, metadata, version,
	created_at, updated_at, authorized_at, captured_at, voided_at, failed_at
`

func (r *IntentRepo) CreateIntent(ctx context.Context, i domain.PaymentIntent) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO payment_intents (
			id, tenant_id, principal_id, order_id, status, amount_minor, captured_minor, refunded_minor,
			currency, method_type, payment_method_id, provider, provider_intent_ref, idempotency_key,
			fraud_score, fraud_decision, failure_reason, metadata, version,
			created_at, updated_at, authorized_at, captured_at, voided_at, failed_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,
			$9,$10,$11,$12,$13,$14,
			$15,$16,$17,$18,$19,
			$20,$21,$22,$23,$24,$25
		)`,
		i.ID, i.TenantID, i.PrincipalID, i.OrderID, string(i.Status), i.AmountMinor, i.CapturedMinor, i.RefundedMinor,
		i.Currency, string(i.MethodType), nullUUID(i.PaymentMethodID), i.Provider, i.ProviderIntentRef, i.IdempotencyKey,
		i.FraudScore, i.FraudDecision, i.FailureReason, JSONMap(i.Metadata), i.Version,
		i.CreatedAt, i.UpdatedAt, nullTime(i.AuthorizedAt), nullTime(i.CapturedAt), nullTime(i.VoidedAt), nullTime(i.FailedAt),
	)
	return mapUniqueViolation(err)
}

func (r *IntentRepo) UpdateIntent(ctx context.Context, i domain.PaymentIntent) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE payment_intents SET
			principal_id=$1, order_id=$2, status=$3, amount_minor=$4, captured_minor=$5, refunded_minor=$6,
			currency=$7, method_type=$8, payment_method_id=$9, provider=$10, provider_intent_ref=$11,
			fraud_score=$12, fraud_decision=$13, failure_reason=$14, metadata=$15, version=$16,
			updated_at=$17, authorized_at=$18, captured_at=$19, voided_at=$20, failed_at=$21
		WHERE id=$22 AND tenant_id=$23`,
		i.PrincipalID, i.OrderID, string(i.Status), i.AmountMinor, i.CapturedMinor, i.RefundedMinor,
		i.Currency, string(i.MethodType), nullUUID(i.PaymentMethodID), i.Provider, i.ProviderIntentRef,
		i.FraudScore, i.FraudDecision, i.FailureReason, JSONMap(i.Metadata), i.Version,
		i.UpdatedAt, nullTime(i.AuthorizedAt), nullTime(i.CapturedAt), nullTime(i.VoidedAt), nullTime(i.FailedAt),
		i.ID, i.TenantID,
	)
	if err != nil {
		return mapUniqueViolation(err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *IntentRepo) GetIntent(ctx context.Context, tenantID, id uuid.UUID) (domain.PaymentIntent, error) {
	return r.scanIntent(ctx, r.DB.QueryRowContext(ctx, `
		SELECT `+intentColumns+` FROM payment_intents WHERE id=$1 AND tenant_id=$2`, id, tenantID))
}

func (r *IntentRepo) GetByIdempotencyKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.PaymentIntent, error) {
	return r.scanIntent(ctx, r.DB.QueryRowContext(ctx, `
		SELECT `+intentColumns+` FROM payment_intents WHERE tenant_id=$1 AND idempotency_key=$2`, tenantID, key))
}

func (r *IntentRepo) ListIntents(ctx context.Context, f ports.IntentFilter) ([]domain.PaymentIntent, int, error) {
	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}

	where := []string{"tenant_id = $1"}
	args := []any{f.TenantID}
	argN := 2

	if f.PrincipalID != nil {
		where = append(where, fmt.Sprintf("principal_id = $%d", argN))
		args = append(args, *f.PrincipalID)
		argN++
	}
	if f.Status != nil {
		where = append(where, fmt.Sprintf("status = $%d", argN))
		args = append(args, string(*f.Status))
		argN++
	}
	if f.OrderID != "" {
		where = append(where, fmt.Sprintf("order_id = $%d", argN))
		args = append(args, f.OrderID)
		argN++
	}
	if q := strings.TrimSpace(f.Query); q != "" {
		where = append(where, fmt.Sprintf("(order_id ILIKE $%d OR idempotency_key ILIKE $%d OR provider_intent_ref ILIKE $%d)", argN, argN, argN))
		args = append(args, "%"+q+"%")
		argN++
	}

	whereSQL := strings.Join(where, " AND ")

	var total int
	if err := r.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM payment_intents WHERE `+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	rows, err := r.DB.QueryContext(ctx, `
		SELECT `+intentColumns+` FROM payment_intents
		WHERE `+whereSQL+`
		ORDER BY created_at DESC
		LIMIT $`+fmt.Sprint(argN)+` OFFSET $`+fmt.Sprint(argN+1), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]domain.PaymentIntent, 0)
	for rows.Next() {
		i, err := scanIntentRow(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, i)
	}
	return out, total, rows.Err()
}

func (r *IntentRepo) CreateAttempt(ctx context.Context, a domain.PaymentAttempt) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO payment_attempts (
			id, intent_id, tenant_id, kind, status, provider, provider_ref,
			amount_minor, currency, error_code, error_message, idempotency_key, is_failover, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		a.ID, a.IntentID, a.TenantID, string(a.Kind), string(a.Status), a.Provider, a.ProviderRef,
		a.AmountMinor, a.Currency, a.ErrorCode, a.ErrorMessage, a.IdempotencyKey, a.IsFailover, a.CreatedAt,
	)
	return err
}

func (r *IntentRepo) ListAttempts(ctx context.Context, tenantID, intentID uuid.UUID) ([]domain.PaymentAttempt, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, intent_id, tenant_id, kind, status, provider, provider_ref,
			amount_minor, currency, error_code, error_message, idempotency_key, is_failover, created_at
		FROM payment_attempts
		WHERE tenant_id=$1 AND intent_id=$2
		ORDER BY created_at ASC`, tenantID, intentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.PaymentAttempt, 0)
	for rows.Next() {
		var a domain.PaymentAttempt
		var kind, status string
		if err := rows.Scan(
			&a.ID, &a.IntentID, &a.TenantID, &kind, &status, &a.Provider, &a.ProviderRef,
			&a.AmountMinor, &a.Currency, &a.ErrorCode, &a.ErrorMessage, &a.IdempotencyKey, &a.IsFailover, &a.CreatedAt,
		); err != nil {
			return nil, err
		}
		a.Kind = domain.AttemptKind(kind)
		a.Status = domain.AttemptStatus(status)
		a.CreatedAt = a.CreatedAt.UTC()
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *IntentRepo) CreateMethod(ctx context.Context, m domain.PaymentMethod) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO payment_methods (
			id, tenant_id, principal_id, method_type, token, last4, brand,
			exp_month, exp_year, provider, active, metadata, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		m.ID, m.TenantID, m.PrincipalID, string(m.MethodType), m.Token, m.Last4, m.Brand,
		m.ExpMonth, m.ExpYear, m.Provider, m.Active, JSONMap(m.Metadata), m.CreatedAt, m.UpdatedAt,
	)
	return mapUniqueViolation(err)
}

func (r *IntentRepo) GetMethod(ctx context.Context, tenantID, id uuid.UUID) (domain.PaymentMethod, error) {
	var m domain.PaymentMethod
	var methodType string
	var meta JSONMap
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, principal_id, method_type, token, last4, brand,
			exp_month, exp_year, provider, active, metadata, created_at, updated_at
		FROM payment_methods WHERE id=$1 AND tenant_id=$2`, id, tenantID).Scan(
		&m.ID, &m.TenantID, &m.PrincipalID, &methodType, &m.Token, &m.Last4, &m.Brand,
		&m.ExpMonth, &m.ExpYear, &m.Provider, &m.Active, &meta, &m.CreatedAt, &m.UpdatedAt,
	)
	if isNoRows(err) {
		return domain.PaymentMethod{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.PaymentMethod{}, err
	}
	m.MethodType = domain.PaymentMethodType(methodType)
	m.Metadata = map[string]any(meta)
	if m.Metadata == nil {
		m.Metadata = map[string]any{}
	}
	m.CreatedAt = m.CreatedAt.UTC()
	m.UpdatedAt = m.UpdatedAt.UTC()
	return m, nil
}

func (r *IntentRepo) ListMethods(ctx context.Context, tenantID, principalID uuid.UUID) ([]domain.PaymentMethod, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, principal_id, method_type, token, last4, brand,
			exp_month, exp_year, provider, active, metadata, created_at, updated_at
		FROM payment_methods
		WHERE tenant_id=$1 AND principal_id=$2
		ORDER BY created_at DESC`, tenantID, principalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.PaymentMethod, 0)
	for rows.Next() {
		var m domain.PaymentMethod
		var methodType string
		var meta JSONMap
		if err := rows.Scan(
			&m.ID, &m.TenantID, &m.PrincipalID, &methodType, &m.Token, &m.Last4, &m.Brand,
			&m.ExpMonth, &m.ExpYear, &m.Provider, &m.Active, &meta, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, err
		}
		m.MethodType = domain.PaymentMethodType(methodType)
		m.Metadata = map[string]any(meta)
		if m.Metadata == nil {
			m.Metadata = map[string]any{}
		}
		m.CreatedAt = m.CreatedAt.UTC()
		m.UpdatedAt = m.UpdatedAt.UTC()
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *IntentRepo) CreateRefund(ctx context.Context, rf domain.Refund) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO refunds (
			id, intent_id, tenant_id, amount_minor, currency, status, provider, provider_ref,
			reason, idempotency_key, created_at, completed_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		rf.ID, rf.IntentID, rf.TenantID, rf.AmountMinor, rf.Currency, string(rf.Status), rf.Provider, rf.ProviderRef,
		rf.Reason, rf.IdempotencyKey, rf.CreatedAt, nullTime(rf.CompletedAt),
	)
	return mapUniqueViolation(err)
}

func (r *IntentRepo) UpdateRefund(ctx context.Context, rf domain.Refund) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE refunds SET
			amount_minor=$1, currency=$2, status=$3, provider=$4, provider_ref=$5,
			reason=$6, completed_at=$7
		WHERE id=$8 AND tenant_id=$9`,
		rf.AmountMinor, rf.Currency, string(rf.Status), rf.Provider, rf.ProviderRef,
		rf.Reason, nullTime(rf.CompletedAt), rf.ID, rf.TenantID,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *IntentRepo) GetRefundByIdempotency(ctx context.Context, tenantID uuid.UUID, key string) (domain.Refund, error) {
	var rf domain.Refund
	var status string
	var completed sql.NullTime
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, intent_id, tenant_id, amount_minor, currency, status, provider, provider_ref,
			reason, idempotency_key, created_at, completed_at
		FROM refunds WHERE tenant_id=$1 AND idempotency_key=$2`, tenantID, key).Scan(
		&rf.ID, &rf.IntentID, &rf.TenantID, &rf.AmountMinor, &rf.Currency, &status, &rf.Provider, &rf.ProviderRef,
		&rf.Reason, &rf.IdempotencyKey, &rf.CreatedAt, &completed,
	)
	if isNoRows(err) {
		return domain.Refund{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Refund{}, err
	}
	rf.Status = domain.RefundStatus(status)
	rf.CompletedAt = scanNullTime(completed)
	rf.CreatedAt = rf.CreatedAt.UTC()
	return rf, nil
}

func (r *IntentRepo) ListRefunds(ctx context.Context, tenantID, intentID uuid.UUID) ([]domain.Refund, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, intent_id, tenant_id, amount_minor, currency, status, provider, provider_ref,
			reason, idempotency_key, created_at, completed_at
		FROM refunds
		WHERE tenant_id=$1 AND intent_id=$2
		ORDER BY created_at ASC`, tenantID, intentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.Refund, 0)
	for rows.Next() {
		var rf domain.Refund
		var status string
		var completed sql.NullTime
		if err := rows.Scan(
			&rf.ID, &rf.IntentID, &rf.TenantID, &rf.AmountMinor, &rf.Currency, &status, &rf.Provider, &rf.ProviderRef,
			&rf.Reason, &rf.IdempotencyKey, &rf.CreatedAt, &completed,
		); err != nil {
			return nil, err
		}
		rf.Status = domain.RefundStatus(status)
		rf.CompletedAt = scanNullTime(completed)
		rf.CreatedAt = rf.CreatedAt.UTC()
		out = append(out, rf)
	}
	return out, rows.Err()
}

func (r *IntentRepo) CreateChargeback(ctx context.Context, c domain.Chargeback) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO chargebacks (
			id, intent_id, tenant_id, amount_minor, currency, status, provider, provider_ref,
			reason_code, reason, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		c.ID, c.IntentID, c.TenantID, c.AmountMinor, c.Currency, string(c.Status), c.Provider, c.ProviderRef,
		c.ReasonCode, c.Reason, c.CreatedAt, c.UpdatedAt,
	)
	return err
}

func (r *IntentRepo) ListChargebacks(ctx context.Context, tenantID uuid.UUID, intentID *uuid.UUID) ([]domain.Chargeback, error) {
	var (
		rows *sql.Rows
		err  error
	)
	if intentID != nil {
		rows, err = r.DB.QueryContext(ctx, `
			SELECT id, intent_id, tenant_id, amount_minor, currency, status, provider, provider_ref,
				reason_code, reason, created_at, updated_at
			FROM chargebacks
			WHERE tenant_id=$1 AND intent_id=$2
			ORDER BY created_at DESC`, tenantID, *intentID)
	} else {
		rows, err = r.DB.QueryContext(ctx, `
			SELECT id, intent_id, tenant_id, amount_minor, currency, status, provider, provider_ref,
				reason_code, reason, created_at, updated_at
			FROM chargebacks
			WHERE tenant_id=$1
			ORDER BY created_at DESC`, tenantID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.Chargeback, 0)
	for rows.Next() {
		var c domain.Chargeback
		var status string
		if err := rows.Scan(
			&c.ID, &c.IntentID, &c.TenantID, &c.AmountMinor, &c.Currency, &status, &c.Provider, &c.ProviderRef,
			&c.ReasonCode, &c.Reason, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		c.Status = domain.ChargebackStatus(status)
		c.CreatedAt = c.CreatedAt.UTC()
		c.UpdatedAt = c.UpdatedAt.UTC()
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *IntentRepo) UpsertRoute(ctx context.Context, route domain.ProviderRoute) error {
	var existingID uuid.UUID
	var createdAt sql.NullTime
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, created_at FROM provider_routes
		WHERE tenant_id=$1 AND method_type=$2 AND currency=$3
		LIMIT 1`, route.TenantID, string(route.MethodType), route.Currency).Scan(&existingID, &createdAt)
	if err != nil && !isNoRows(err) {
		return err
	}
	if err == nil {
		_, err = r.DB.ExecContext(ctx, `
			UPDATE provider_routes SET
				providers=$1, active=$2, priority=$3, updated_at=$4
			WHERE id=$5`,
			TextArray(route.Providers), route.Active, route.Priority, route.UpdatedAt, existingID,
		)
		return err
	}
	_, err = r.DB.ExecContext(ctx, `
		INSERT INTO provider_routes (
			id, tenant_id, method_type, currency, providers, active, priority, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		route.ID, route.TenantID, string(route.MethodType), route.Currency,
		TextArray(route.Providers), route.Active, route.Priority, route.CreatedAt, route.UpdatedAt,
	)
	return err
}

func (r *IntentRepo) ListRoutes(ctx context.Context, tenantID uuid.UUID, method domain.PaymentMethodType, currency string) ([]domain.ProviderRoute, error) {
	where := []string{"tenant_id = $1", "active = true"}
	args := []any{tenantID}
	argN := 2
	if method != "" {
		where = append(where, fmt.Sprintf("method_type = $%d", argN))
		args = append(args, string(method))
		argN++
	}
	if currency != "" {
		where = append(where, fmt.Sprintf("(currency = '' OR currency = $%d)", argN))
		args = append(args, currency)
		argN++
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, method_type, currency, providers, active, priority, created_at, updated_at
		FROM provider_routes
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY priority ASC`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.ProviderRoute, 0)
	for rows.Next() {
		var route domain.ProviderRoute
		var methodType string
		var providers TextArray
		if err := rows.Scan(
			&route.ID, &route.TenantID, &methodType, &route.Currency, &providers,
			&route.Active, &route.Priority, &route.CreatedAt, &route.UpdatedAt,
		); err != nil {
			return nil, err
		}
		route.MethodType = domain.PaymentMethodType(methodType)
		route.Providers = []string(providers)
		if route.Providers == nil {
			route.Providers = []string{}
		}
		route.CreatedAt = route.CreatedAt.UTC()
		route.UpdatedAt = route.UpdatedAt.UTC()
		out = append(out, route)
	}
	return out, rows.Err()
}

func (r *IntentRepo) CreateFraudScore(ctx context.Context, f domain.FraudScore) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO fraud_scores (
			id, intent_id, tenant_id, score, decision, reasons, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		f.ID, f.IntentID, f.TenantID, f.Score, string(f.Decision), TextArray(f.Reasons), f.CreatedAt,
	)
	return err
}

func (r *IntentRepo) CreateAudit(ctx context.Context, a domain.AuditEntry) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO payment_audit (
			id, tenant_id, intent_id, action, actor_id, actor_type,
			amount_minor, currency, detail, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		a.ID, a.TenantID, nullUUID(a.IntentID), a.Action, nullUUID(a.ActorID), a.ActorType,
		a.AmountMinor, a.Currency, JSONMap(a.Detail), a.CreatedAt,
	)
	return err
}

type scannable interface {
	Scan(dest ...any) error
}

func (r *IntentRepo) scanIntent(ctx context.Context, row scannable) (domain.PaymentIntent, error) {
	_ = ctx
	i, err := scanIntentRow(row)
	if isNoRows(err) {
		return domain.PaymentIntent{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.PaymentIntent{}, err
	}
	return i, nil
}

func scanIntentRow(row scannable) (domain.PaymentIntent, error) {
	var i domain.PaymentIntent
	var status, methodType string
	var methodID uuid.NullUUID
	var meta JSONMap
	var authorized, captured, voided, failed sql.NullTime
	err := row.Scan(
		&i.ID, &i.TenantID, &i.PrincipalID, &i.OrderID, &status, &i.AmountMinor, &i.CapturedMinor, &i.RefundedMinor,
		&i.Currency, &methodType, &methodID, &i.Provider, &i.ProviderIntentRef, &i.IdempotencyKey,
		&i.FraudScore, &i.FraudDecision, &i.FailureReason, &meta, &i.Version,
		&i.CreatedAt, &i.UpdatedAt, &authorized, &captured, &voided, &failed,
	)
	if err != nil {
		return domain.PaymentIntent{}, err
	}
	i.Status = domain.IntentStatus(status)
	i.MethodType = domain.PaymentMethodType(methodType)
	i.PaymentMethodID = scanNullUUID(methodID)
	i.Metadata = map[string]any(meta)
	if i.Metadata == nil {
		i.Metadata = map[string]any{}
	}
	i.AuthorizedAt = scanNullTime(authorized)
	i.CapturedAt = scanNullTime(captured)
	i.VoidedAt = scanNullTime(voided)
	i.FailedAt = scanNullTime(failed)
	i.CreatedAt = i.CreatedAt.UTC()
	i.UpdatedAt = i.UpdatedAt.UTC()
	return i, nil
}
