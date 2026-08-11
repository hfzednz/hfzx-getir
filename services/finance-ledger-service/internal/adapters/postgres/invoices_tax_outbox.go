package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/finance-ledger-service/internal/app/ports"
	"github.com/nexora/finance-ledger-service/internal/domain"
)

// InvoiceRepo persists invoices and credit notes.
type InvoiceRepo struct{ DB *sql.DB }

func (r *InvoiceRepo) Create(ctx context.Context, inv domain.Invoice) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	_, err = tx.ExecContext(ctx, `
		INSERT INTO invoices (
			id, tenant_id, status, currency, counterparty_ref, external_ref, idempotency_key,
			subtotal_minor, tax_minor, total_minor, issued_at, created_at, updated_at, version
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		inv.ID, inv.TenantID, string(inv.Status), inv.Currency, inv.CounterpartyRef, inv.ExternalRef,
		nullString(inv.IdempotencyKey), inv.SubtotalMinor, inv.TaxMinor, inv.TotalMinor, nullTime(inv.IssuedAt),
		inv.CreatedAt.UTC(), inv.UpdatedAt.UTC(), inv.Version)
	if err != nil {
		return mapUniqueViolation(err)
	}
	if err := r.insertLines(ctx, tx, inv); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *InvoiceRepo) Update(ctx context.Context, inv domain.Invoice) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	res, err := tx.ExecContext(ctx, `
		UPDATE invoices SET status=$3, currency=$4, counterparty_ref=$5, external_ref=$6, idempotency_key=$7,
			subtotal_minor=$8, tax_minor=$9, total_minor=$10, issued_at=$11, updated_at=$12, version=$13
		WHERE id=$1 AND tenant_id=$2`,
		inv.ID, inv.TenantID, string(inv.Status), inv.Currency, inv.CounterpartyRef, inv.ExternalRef,
		nullString(inv.IdempotencyKey), inv.SubtotalMinor, inv.TaxMinor, inv.TotalMinor, nullTime(inv.IssuedAt),
		inv.UpdatedAt.UTC(), inv.Version)
	if err != nil {
		return mapUniqueViolation(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM invoice_lines WHERE invoice_id=$1 AND tenant_id=$2`, inv.ID, inv.TenantID); err != nil {
		return err
	}
	if err := r.insertLines(ctx, tx, inv); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *InvoiceRepo) insertLines(ctx context.Context, tx *sql.Tx, inv domain.Invoice) error {
	for _, l := range inv.Lines {
		id := l.ID
		if id == uuid.Nil {
			id = uuid.New()
		}
		_, err := tx.ExecContext(ctx, `
			INSERT INTO invoice_lines (
				id, invoice_id, tenant_id, description, qty, unit_minor, tax_minor, total_minor, tax_code
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
			id, inv.ID, inv.TenantID, l.Description, l.Qty, l.UnitMinor, l.TaxMinor, l.TotalMinor, l.TaxCode)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *InvoiceRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Invoice, error) {
	inv, err := r.scanHeader(r.DB.QueryRowContext(ctx, invoiceSelect+` WHERE id=$1 AND tenant_id=$2`, id, tenantID))
	if isNoRows(err) {
		return domain.Invoice{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Invoice{}, err
	}
	lines, err := r.loadLines(ctx, tenantID, id)
	if err != nil {
		return domain.Invoice{}, err
	}
	inv.Lines = lines
	return inv, nil
}

func (r *InvoiceRepo) GetByIdempotencyKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.Invoice, error) {
	inv, err := r.scanHeader(r.DB.QueryRowContext(ctx, invoiceSelect+` WHERE tenant_id=$1 AND idempotency_key=$2`, tenantID, key))
	if isNoRows(err) {
		return domain.Invoice{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Invoice{}, err
	}
	lines, err := r.loadLines(ctx, tenantID, inv.ID)
	if err != nil {
		return domain.Invoice{}, err
	}
	inv.Lines = lines
	return inv, nil
}

func (r *InvoiceRepo) CreateCreditNote(ctx context.Context, cn domain.CreditNote) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO credit_notes (
			id, tenant_id, invoice_id, currency, amount_minor, reason, idempotency_key, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		cn.ID, cn.TenantID, cn.InvoiceID, cn.Currency, cn.AmountMinor, cn.Reason, nullString(cn.IdempotencyKey), cn.CreatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *InvoiceRepo) GetCreditNoteByIdempotencyKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.CreditNote, error) {
	var cn domain.CreditNote
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, invoice_id, currency, amount_minor, reason, COALESCE(idempotency_key,''), created_at
		FROM credit_notes WHERE tenant_id=$1 AND idempotency_key=$2`, tenantID, key).Scan(
		&cn.ID, &cn.TenantID, &cn.InvoiceID, &cn.Currency, &cn.AmountMinor, &cn.Reason, &cn.IdempotencyKey, &cn.CreatedAt)
	if isNoRows(err) {
		return domain.CreditNote{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.CreditNote{}, err
	}
	cn.CreatedAt = cn.CreatedAt.UTC()
	return cn, nil
}

const invoiceSelect = `
	SELECT id, tenant_id, status, currency, counterparty_ref, external_ref, COALESCE(idempotency_key,''),
		subtotal_minor, tax_minor, total_minor, issued_at, created_at, updated_at, version FROM invoices`

func (r *InvoiceRepo) scanHeader(row *sql.Row) (domain.Invoice, error) {
	var inv domain.Invoice
	var status string
	var issued sql.NullTime
	err := row.Scan(
		&inv.ID, &inv.TenantID, &status, &inv.Currency, &inv.CounterpartyRef, &inv.ExternalRef, &inv.IdempotencyKey,
		&inv.SubtotalMinor, &inv.TaxMinor, &inv.TotalMinor, &issued, &inv.CreatedAt, &inv.UpdatedAt, &inv.Version)
	if err != nil {
		return domain.Invoice{}, err
	}
	inv.Status = domain.InvoiceStatus(status)
	inv.IssuedAt = scanNullTime(issued)
	inv.CreatedAt = inv.CreatedAt.UTC()
	inv.UpdatedAt = inv.UpdatedAt.UTC()
	return inv, nil
}

func (r *InvoiceRepo) loadLines(ctx context.Context, tenantID, invoiceID uuid.UUID) ([]domain.InvoiceLine, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, description, qty, unit_minor, tax_minor, total_minor, tax_code
		FROM invoice_lines WHERE tenant_id=$1 AND invoice_id=$2`, tenantID, invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.InvoiceLine{}
	for rows.Next() {
		var l domain.InvoiceLine
		if err := rows.Scan(&l.ID, &l.Description, &l.Qty, &l.UnitMinor, &l.TaxMinor, &l.TotalMinor, &l.TaxCode); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

// TaxRuleRepo persists tax rules.
type TaxRuleRepo struct{ DB *sql.DB }

func (r *TaxRuleRepo) Upsert(ctx context.Context, rule domain.TaxRule) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO tax_rules (id, tenant_id, code, name, rate_bps, currency, active, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (tenant_id, code) DO UPDATE SET
			name=EXCLUDED.name, rate_bps=EXCLUDED.rate_bps, currency=EXCLUDED.currency,
			active=EXCLUDED.active, updated_at=EXCLUDED.updated_at`,
		rule.ID, rule.TenantID, rule.Code, rule.Name, rule.RateBps, rule.Currency, rule.Active,
		rule.CreatedAt.UTC(), rule.UpdatedAt.UTC())
	return err
}

func (r *TaxRuleRepo) GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (domain.TaxRule, error) {
	var rule domain.TaxRule
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, code, name, rate_bps, currency, active, created_at, updated_at
		FROM tax_rules WHERE tenant_id=$1 AND code=$2`, tenantID, code).Scan(
		&rule.ID, &rule.TenantID, &rule.Code, &rule.Name, &rule.RateBps, &rule.Currency, &rule.Active,
		&rule.CreatedAt, &rule.UpdatedAt)
	if isNoRows(err) {
		return domain.TaxRule{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.TaxRule{}, err
	}
	rule.CreatedAt = rule.CreatedAt.UTC()
	rule.UpdatedAt = rule.UpdatedAt.UTC()
	return rule, nil
}

// EventRepo persists ledger timeline events.
type EventRepo struct{ DB *sql.DB }

func (r *EventRepo) Append(ctx context.Context, e domain.LedgerEvent) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO ledger_events (
			id, entity_id, tenant_id, event_type, payload, actor_id, actor_type, occurred_at, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		e.ID, e.EntityID, e.TenantID, e.Type, JSONMap(e.Payload), nullUUID(e.ActorID), e.ActorType,
		e.OccurredAt.UTC(), e.CreatedAt.UTC())
	return err
}

func (r *EventRepo) ListByEntity(ctx context.Context, tenantID, entityID uuid.UUID) ([]domain.LedgerEvent, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, entity_id, tenant_id, event_type, payload, actor_id, actor_type, occurred_at, created_at
		FROM ledger_events WHERE tenant_id=$1 AND entity_id=$2 ORDER BY occurred_at ASC`, tenantID, entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.LedgerEvent{}
	for rows.Next() {
		var e domain.LedgerEvent
		var payload JSONMap
		var actor uuid.NullUUID
		if err := rows.Scan(
			&e.ID, &e.EntityID, &e.TenantID, &e.Type, &payload, &actor, &e.ActorType, &e.OccurredAt, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.Payload = map[string]any(payload)
		e.ActorID = scanNullUUID(actor)
		e.OccurredAt = e.OccurredAt.UTC()
		e.CreatedAt = e.CreatedAt.UTC()
		out = append(out, e)
	}
	return out, rows.Err()
}

// OutboxRepo persists transactional outbox rows.
type OutboxRepo struct{ DB *sql.DB }

func (r *OutboxRepo) Enqueue(ctx context.Context, m domain.OutboxMessage) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO ledger_outbox (
			id, tenant_id, entity_id, topic, message_key, payload, status, attempts, last_error,
			created_at, updated_at, published_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		m.ID, m.TenantID, m.EntityID, m.Topic, m.Key, JSONMap(m.Payload), string(m.Status), m.Attempts, m.LastError,
		m.CreatedAt.UTC(), m.UpdatedAt.UTC(), nullTime(m.PublishedAt))
	return err
}

func (r *OutboxRepo) Update(ctx context.Context, m domain.OutboxMessage) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE ledger_outbox SET
			topic=$2, message_key=$3, payload=$4, status=$5, attempts=$6, last_error=$7,
			updated_at=$8, published_at=$9
		WHERE id=$1`,
		m.ID, m.Topic, m.Key, JSONMap(m.Payload), string(m.Status), m.Attempts, m.LastError,
		m.UpdatedAt.UTC(), nullTime(m.PublishedAt))
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *OutboxRepo) ListPending(ctx context.Context, limit int) ([]domain.OutboxMessage, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, entity_id, topic, message_key, payload, status, attempts, last_error,
			created_at, updated_at, published_at
		FROM ledger_outbox WHERE status='pending' ORDER BY created_at ASC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.OutboxMessage{}
	for rows.Next() {
		var m domain.OutboxMessage
		var status string
		var payload JSONMap
		var published sql.NullTime
		if err := rows.Scan(
			&m.ID, &m.TenantID, &m.EntityID, &m.Topic, &m.Key, &payload, &status, &m.Attempts, &m.LastError,
			&m.CreatedAt, &m.UpdatedAt, &published); err != nil {
			return nil, err
		}
		m.Status = domain.OutboxStatus(status)
		m.Payload = map[string]any(payload)
		m.CreatedAt = m.CreatedAt.UTC()
		m.UpdatedAt = m.UpdatedAt.UTC()
		m.PublishedAt = scanNullTime(published)
		out = append(out, m)
	}
	return out, rows.Err()
}

// Repos groups finance-ledger persistence adapters.
type Repos struct {
	Accounts *AccountRepo
	Journals *JournalRepo
	Invoices *InvoiceRepo
	TaxRules *TaxRuleRepo
	Events   *EventRepo
	Outbox   *OutboxRepo
}

// NewRepos constructs postgres-backed repositories.
func NewRepos(db *sql.DB) *Repos {
	return &Repos{
		Accounts: &AccountRepo{DB: db},
		Journals: &JournalRepo{DB: db},
		Invoices: &InvoiceRepo{DB: db},
		TaxRules: &TaxRuleRepo{DB: db},
		Events:   &EventRepo{DB: db},
		Outbox:   &OutboxRepo{DB: db},
	}
}

var (
	_ ports.InvoiceRepository = (*InvoiceRepo)(nil)
	_ ports.TaxRuleRepository = (*TaxRuleRepo)(nil)
	_ ports.EventStore        = (*EventRepo)(nil)
	_ ports.OutboxRepository  = (*OutboxRepo)(nil)
)
