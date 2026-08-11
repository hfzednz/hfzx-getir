package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/settlement-service/internal/app/ports"
	"github.com/nexora/settlement-service/internal/domain"
)

// BatchRepo persists settlement batches and related rows.
type BatchRepo struct{ DB *sql.DB }

func (r *BatchRepo) Create(ctx context.Context, b domain.SettlementBatch) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	_, err = tx.ExecContext(ctx, `
		INSERT INTO settlement_batches (
			id, tenant_id, status, currency, period_start, period_end, description, idempotency_key,
			total_minor, submitted_by, submitted_at, approved_by, approved_at, completed_at, failed_at,
			failure_reason, created_at, updated_at, version
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)`,
		b.ID, b.TenantID, string(b.Status), b.Currency, b.PeriodStart.UTC(), b.PeriodEnd.UTC(), b.Description,
		nullString(b.IdempotencyKey), b.TotalMinor, nullUUID(b.SubmittedBy), nullTime(b.SubmittedAt),
		nullUUID(b.ApprovedBy), nullTime(b.ApprovedAt), nullTime(b.CompletedAt), nullTime(b.FailedAt),
		b.FailureReason, b.CreatedAt.UTC(), b.UpdatedAt.UTC(), b.Version)
	if err != nil {
		return mapUniqueViolation(err)
	}
	if err := r.replaceLines(ctx, tx, b); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *BatchRepo) Update(ctx context.Context, b domain.SettlementBatch) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	res, err := tx.ExecContext(ctx, `
		UPDATE settlement_batches SET
			status=$3, currency=$4, period_start=$5, period_end=$6, description=$7, idempotency_key=$8,
			total_minor=$9, submitted_by=$10, submitted_at=$11, approved_by=$12, approved_at=$13,
			completed_at=$14, failed_at=$15, failure_reason=$16, updated_at=$17, version=$18
		WHERE id=$1 AND tenant_id=$2 AND version <= $18`,
		b.ID, b.TenantID, string(b.Status), b.Currency, b.PeriodStart.UTC(), b.PeriodEnd.UTC(), b.Description,
		nullString(b.IdempotencyKey), b.TotalMinor, nullUUID(b.SubmittedBy), nullTime(b.SubmittedAt),
		nullUUID(b.ApprovedBy), nullTime(b.ApprovedAt), nullTime(b.CompletedAt), nullTime(b.FailedAt),
		b.FailureReason, b.UpdatedAt.UTC(), b.Version)
	if err != nil {
		return mapUniqueViolation(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM settlement_lines WHERE batch_id=$1 AND tenant_id=$2`, b.ID, b.TenantID); err != nil {
		return err
	}
	if err := r.replaceLines(ctx, tx, b); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *BatchRepo) replaceLines(ctx context.Context, tx *sql.Tx, b domain.SettlementBatch) error {
	for _, l := range b.Lines {
		lineID := l.ID
		if lineID == uuid.Nil {
			lineID = uuid.New()
		}
		_, err := tx.ExecContext(ctx, `
			INSERT INTO settlement_lines (
				id, batch_id, tenant_id, payee_type, payee_ref, amount_minor, currency, external_ref, memo
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
			lineID, b.ID, b.TenantID, string(l.PayeeType), l.PayeeRef, l.AmountMinor, l.Currency, l.ExternalRef, l.Memo)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *BatchRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.SettlementBatch, error) {
	b, err := r.scanBatch(r.DB.QueryRowContext(ctx, batchSelect+` WHERE id=$1 AND tenant_id=$2`, id, tenantID))
	if isNoRows(err) {
		return domain.SettlementBatch{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.SettlementBatch{}, err
	}
	lines, err := r.loadLines(ctx, tenantID, id)
	if err != nil {
		return domain.SettlementBatch{}, err
	}
	b.Lines = lines
	return b, nil
}

func (r *BatchRepo) GetByIdempotencyKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.SettlementBatch, error) {
	b, err := r.scanBatch(r.DB.QueryRowContext(ctx, batchSelect+` WHERE tenant_id=$1 AND idempotency_key=$2`, tenantID, key))
	if isNoRows(err) {
		return domain.SettlementBatch{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.SettlementBatch{}, err
	}
	lines, err := r.loadLines(ctx, tenantID, b.ID)
	if err != nil {
		return domain.SettlementBatch{}, err
	}
	b.Lines = lines
	return b, nil
}

func (r *BatchRepo) List(ctx context.Context, f ports.BatchFilter) ([]domain.SettlementBatch, int, error) {
	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}
	var total int
	args := []any{f.TenantID}
	where := `WHERE tenant_id=$1`
	if f.Status != nil {
		where += ` AND status=$2`
		args = append(args, string(*f.Status))
	}
	if err := r.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM settlement_batches `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	q := batchSelect + ` ` + where + ` ORDER BY created_at DESC LIMIT $` + itoa(len(args)+1) + ` OFFSET $` + itoa(len(args)+2)
	args = append(args, limit, offset)
	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.SettlementBatch{}
	for rows.Next() {
		b, err := scanBatchRow(rows)
		if err != nil {
			return nil, 0, err
		}
		lines, err := r.loadLines(ctx, f.TenantID, b.ID)
		if err != nil {
			return nil, 0, err
		}
		b.Lines = lines
		out = append(out, b)
	}
	return out, total, rows.Err()
}

func itoa(n int) string {
	const digits = "0123456789"
	if n < 10 {
		return digits[n : n+1]
	}
	return itoa(n/10) + digits[n%10:n%10+1]
}

func (r *BatchRepo) SavePayout(ctx context.Context, p domain.PayoutInstruction) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO payout_instructions (
			id, batch_id, line_id, tenant_id, payee_type, payee_ref, amount_minor, currency,
			status, provider_ref, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (id) DO UPDATE SET
			status=EXCLUDED.status, provider_ref=EXCLUDED.provider_ref, updated_at=EXCLUDED.updated_at`,
		p.ID, p.BatchID, p.LineID, p.TenantID, string(p.PayeeType), p.PayeeRef, p.AmountMinor, p.Currency,
		p.Status, p.ProviderRef, p.CreatedAt.UTC(), p.UpdatedAt.UTC())
	return err
}

func (r *BatchRepo) ListPayouts(ctx context.Context, tenantID, batchID uuid.UUID) ([]domain.PayoutInstruction, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, batch_id, line_id, tenant_id, payee_type, payee_ref, amount_minor, currency,
			status, provider_ref, created_at, updated_at
		FROM payout_instructions WHERE tenant_id=$1 AND batch_id=$2`, tenantID, batchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.PayoutInstruction{}
	for rows.Next() {
		var p domain.PayoutInstruction
		var payee string
		if err := rows.Scan(
			&p.ID, &p.BatchID, &p.LineID, &p.TenantID, &payee, &p.PayeeRef, &p.AmountMinor, &p.Currency,
			&p.Status, &p.ProviderRef, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		p.PayeeType = domain.PayeeType(payee)
		p.CreatedAt = p.CreatedAt.UTC()
		p.UpdatedAt = p.UpdatedAt.UTC()
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *BatchRepo) SaveReconciliation(ctx context.Context, rec domain.Reconciliation) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO reconciliations (
			id, tenant_id, batch_id, provider_ref, expected_minor, reported_minor, matched, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		rec.ID, rec.TenantID, rec.BatchID, rec.ProviderRef, rec.ExpectedMinor, rec.ReportedMinor, rec.Matched, rec.CreatedAt.UTC())
	return err
}

func (r *BatchRepo) SaveMismatch(ctx context.Context, m domain.Mismatch) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO mismatches (
			id, tenant_id, batch_id, reconcile_id, expected_minor, reported_minor, delta_minor, detail, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		m.ID, m.TenantID, m.BatchID, m.ReconcileID, m.ExpectedMinor, m.ReportedMinor, m.DeltaMinor, m.Detail, m.CreatedAt.UTC())
	return err
}

func (r *BatchRepo) ListMismatches(ctx context.Context, tenantID, batchID uuid.UUID) ([]domain.Mismatch, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, batch_id, reconcile_id, expected_minor, reported_minor, delta_minor, detail, created_at
		FROM mismatches WHERE tenant_id=$1 AND batch_id=$2`, tenantID, batchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Mismatch{}
	for rows.Next() {
		var m domain.Mismatch
		if err := rows.Scan(
			&m.ID, &m.TenantID, &m.BatchID, &m.ReconcileID, &m.ExpectedMinor, &m.ReportedMinor,
			&m.DeltaMinor, &m.Detail, &m.CreatedAt); err != nil {
			return nil, err
		}
		m.CreatedAt = m.CreatedAt.UTC()
		out = append(out, m)
	}
	return out, rows.Err()
}

const batchSelect = `
	SELECT id, tenant_id, status, currency, period_start, period_end, description, COALESCE(idempotency_key,''),
		total_minor, submitted_by, submitted_at, approved_by, approved_at, completed_at, failed_at,
		failure_reason, created_at, updated_at, version
	FROM settlement_batches`

func (r *BatchRepo) scanBatch(row *sql.Row) (domain.SettlementBatch, error) {
	return scanBatchRow(row)
}

type batchScanner interface {
	Scan(dest ...any) error
}

func scanBatchRow(row batchScanner) (domain.SettlementBatch, error) {
	var b domain.SettlementBatch
	var status string
	var submittedBy, approvedBy uuid.NullUUID
	var submittedAt, approvedAt, completedAt, failedAt sql.NullTime
	err := row.Scan(
		&b.ID, &b.TenantID, &status, &b.Currency, &b.PeriodStart, &b.PeriodEnd, &b.Description, &b.IdempotencyKey,
		&b.TotalMinor, &submittedBy, &submittedAt, &approvedBy, &approvedAt, &completedAt, &failedAt,
		&b.FailureReason, &b.CreatedAt, &b.UpdatedAt, &b.Version)
	if err != nil {
		return domain.SettlementBatch{}, err
	}
	b.Status = domain.BatchStatus(status)
	b.SubmittedBy = scanNullUUID(submittedBy)
	b.ApprovedBy = scanNullUUID(approvedBy)
	b.SubmittedAt = scanNullTime(submittedAt)
	b.ApprovedAt = scanNullTime(approvedAt)
	b.CompletedAt = scanNullTime(completedAt)
	b.FailedAt = scanNullTime(failedAt)
	b.PeriodStart = b.PeriodStart.UTC()
	b.PeriodEnd = b.PeriodEnd.UTC()
	b.CreatedAt = b.CreatedAt.UTC()
	b.UpdatedAt = b.UpdatedAt.UTC()
	return b, nil
}

func (r *BatchRepo) loadLines(ctx context.Context, tenantID, batchID uuid.UUID) ([]domain.SettlementLine, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, payee_type, payee_ref, amount_minor, currency, external_ref, memo
		FROM settlement_lines WHERE tenant_id=$1 AND batch_id=$2`, tenantID, batchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.SettlementLine{}
	for rows.Next() {
		var l domain.SettlementLine
		var payee string
		if err := rows.Scan(&l.ID, &payee, &l.PayeeRef, &l.AmountMinor, &l.Currency, &l.ExternalRef, &l.Memo); err != nil {
			return nil, err
		}
		l.PayeeType = domain.PayeeType(payee)
		out = append(out, l)
	}
	return out, rows.Err()
}

var _ ports.BatchRepository = (*BatchRepo)(nil)
