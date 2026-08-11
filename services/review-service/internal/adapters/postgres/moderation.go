package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/review-service/internal/app/ports"
	"github.com/nexora/review-service/internal/domain"
)

// ModerationRepo persists moderation cases.
type ModerationRepo struct{ DB *sql.DB }

func (r *ModerationRepo) Save(ctx context.Context, m domain.ModerationCase) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, `
		DELETE FROM moderation_cases WHERE review_id=$1 AND tenant_id=$2 AND id <> $3`,
		m.ReviewID, m.TenantID, m.ID)
	if err != nil {
		return err
	}

	labels := TextArray(m.Labels)
	signals := TextArray(m.FraudSignals)
	_, err = tx.ExecContext(ctx, `
		INSERT INTO moderation_cases (
		  id, review_id, tenant_id, status, auto_decision, ai_score, labels, fraud_score, fraud_signals,
		  pii_masked, assignee_id, decision_note, decided_by, created_at, updated_at, decided_at
		) VALUES (
		  $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16
		)
		ON CONFLICT (id) DO UPDATE SET
		  status=EXCLUDED.status, auto_decision=EXCLUDED.auto_decision, ai_score=EXCLUDED.ai_score,
		  labels=EXCLUDED.labels, fraud_score=EXCLUDED.fraud_score, fraud_signals=EXCLUDED.fraud_signals,
		  pii_masked=EXCLUDED.pii_masked, assignee_id=EXCLUDED.assignee_id, decision_note=EXCLUDED.decision_note,
		  decided_by=EXCLUDED.decided_by, updated_at=EXCLUDED.updated_at, decided_at=EXCLUDED.decided_at`,
		m.ID, m.ReviewID, m.TenantID, m.Status, m.AutoDecision, m.AIScore, labels, m.FraudScore, signals,
		m.PIIMasked, nullUUID(m.AssigneeID), m.DecisionNote, nullUUID(m.DecidedBy),
		m.CreatedAt.UTC(), m.UpdatedAt.UTC(), nullTime(m.DecidedAt))
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *ModerationRepo) GetByReview(ctx context.Context, tenantID, reviewID uuid.UUID) (domain.ModerationCase, error) {
	var m domain.ModerationCase
	var labels, signals TextArray
	var assignee, decidedBy uuid.NullUUID
	var decidedAt sql.NullTime
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, review_id, tenant_id, status, auto_decision, ai_score, labels, fraud_score, fraud_signals,
		       pii_masked, assignee_id, decision_note, decided_by, created_at, updated_at, decided_at
		FROM moderation_cases
		WHERE tenant_id=$1 AND review_id=$2
		ORDER BY updated_at DESC LIMIT 1`, tenantID, reviewID).Scan(
		&m.ID, &m.ReviewID, &m.TenantID, &m.Status, &m.AutoDecision, &m.AIScore, &labels, &m.FraudScore, &signals,
		&m.PIIMasked, &assignee, &m.DecisionNote, &decidedBy, &m.CreatedAt, &m.UpdatedAt, &decidedAt)
	if err != nil {
		return domain.ModerationCase{}, mapNotFound(err)
	}
	m.Labels = []string(labels)
	m.FraudSignals = []string(signals)
	m.AssigneeID = scanNullUUID(assignee)
	m.DecidedBy = scanNullUUID(decidedBy)
	m.CreatedAt = m.CreatedAt.UTC()
	m.UpdatedAt = m.UpdatedAt.UTC()
	m.DecidedAt = scanNullTime(decidedAt)
	return m, nil
}

func (r *ModerationRepo) ListPending(ctx context.Context, tenantID uuid.UUID, limit int) ([]domain.ModerationCase, error) {
	q := `
		SELECT id, review_id, tenant_id, status, auto_decision, ai_score, labels, fraud_score, fraud_signals,
		       pii_masked, assignee_id, decision_note, decided_by, created_at, updated_at, decided_at
		FROM moderation_cases
		WHERE tenant_id=$1 AND status=$2
		ORDER BY created_at ASC`
	args := []any{tenantID, domain.ModerationPending}
	if limit > 0 {
		q += ` LIMIT $3`
		args = append(args, limit)
	}
	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.ModerationCase, 0)
	for rows.Next() {
		var m domain.ModerationCase
		var labels, signals TextArray
		var assignee, decidedBy uuid.NullUUID
		var decidedAt sql.NullTime
		if err := rows.Scan(
			&m.ID, &m.ReviewID, &m.TenantID, &m.Status, &m.AutoDecision, &m.AIScore, &labels, &m.FraudScore, &signals,
			&m.PIIMasked, &assignee, &m.DecisionNote, &decidedBy, &m.CreatedAt, &m.UpdatedAt, &decidedAt); err != nil {
			return nil, err
		}
		m.Labels = []string(labels)
		m.FraudSignals = []string(signals)
		m.AssigneeID = scanNullUUID(assignee)
		m.DecidedBy = scanNullUUID(decidedBy)
		m.CreatedAt = m.CreatedAt.UTC()
		m.UpdatedAt = m.UpdatedAt.UTC()
		m.DecidedAt = scanNullTime(decidedAt)
		out = append(out, m)
	}
	return out, rows.Err()
}

var _ ports.ModerationRepo = (*ModerationRepo)(nil)
