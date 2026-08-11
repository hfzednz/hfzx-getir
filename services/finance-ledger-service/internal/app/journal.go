package app

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/finance-ledger-service/internal/app/ports"
	"github.com/nexora/finance-ledger-service/internal/domain"
)

// JournalLineInput is a line for PostJournal.
type JournalLineInput struct {
	AccountID   uuid.UUID
	AccountCode string
	DebitMinor  int64
	CreditMinor int64
	Memo        string
}

// PostJournalInput posts a balanced double-entry journal.
type PostJournalInput struct {
	TenantID       uuid.UUID
	Currency       string
	Reference      string
	Description    string
	IdempotencyKey string
	Lines          []JournalLineInput
}

// PostJournal creates and immediately posts a balanced journal.
// Rejects unbalanced journals. Idempotent by IdempotencyKey when set.
func (d *Deps) PostJournal(ctx context.Context, in PostJournalInput) (domain.Journal, error) {
	if in.TenantID == uuid.Nil {
		return domain.Journal{}, fmt.Errorf("%w: tenant_id", domain.ErrInvalidArgument)
	}
	currency := strings.ToUpper(strings.TrimSpace(in.Currency))
	if _, err := domain.NewMoney(0, currency); err != nil {
		return domain.Journal{}, err
	}
	if in.IdempotencyKey != "" {
		if existing, err := d.Journals.GetByIdempotencyKey(ctx, in.TenantID, in.IdempotencyKey); err == nil {
			return existing, nil
		} else if !errors.Is(err, domain.ErrNotFound) {
			return domain.Journal{}, err
		}
	}
	if len(in.Lines) < 2 {
		return domain.Journal{}, fmt.Errorf("%w: at least 2 lines required", domain.ErrInvalidArgument)
	}

	now := d.now()
	lines := make([]domain.JournalLine, 0, len(in.Lines))
	for _, li := range in.Lines {
		accID := li.AccountID
		code := strings.TrimSpace(li.AccountCode)
		if accID == uuid.Nil && code != "" {
			acc, err := d.Accounts.GetByCode(ctx, in.TenantID, code)
			if err != nil {
				return domain.Journal{}, fmt.Errorf("%w: account %q", err, code)
			}
			accID = acc.ID
			code = acc.Code
		} else if accID != uuid.Nil {
			acc, err := d.Accounts.GetByID(ctx, in.TenantID, accID)
			if err != nil {
				return domain.Journal{}, err
			}
			code = acc.Code
		}
		line := domain.JournalLine{
			ID:          d.newID(),
			AccountID:   accID,
			AccountCode: code,
			DebitMinor:  li.DebitMinor,
			CreditMinor: li.CreditMinor,
			Currency:    currency,
			Memo:        li.Memo,
		}
		if err := line.Validate(); err != nil {
			return domain.Journal{}, err
		}
		lines = append(lines, line)
	}

	j := domain.Journal{
		ID:             d.newID(),
		TenantID:       in.TenantID,
		Status:         domain.JournalStatusPosted,
		Currency:       currency,
		Reference:      in.Reference,
		Description:    in.Description,
		IdempotencyKey: in.IdempotencyKey,
		Lines:          lines,
		PostedAt:       &now,
		CreatedAt:      now,
		UpdatedAt:      now,
		Version:        1,
	}
	if err := j.Validate(); err != nil {
		return domain.Journal{}, err
	}
	if err := j.AssertBalanced(); err != nil {
		return domain.Journal{}, err
	}
	if err := d.Journals.Create(ctx, j); err != nil {
		return domain.Journal{}, err
	}
	_ = d.appendEvent(ctx, j.ID, j.TenantID, domain.EventJournalPosted, map[string]any{
		"currency":    j.Currency,
		"debitTotal":  j.DebitTotal(),
		"creditTotal": j.CreditTotal(),
		"reference":   j.Reference,
	})
	return j, nil
}

// ListJournalsInput filters journals.
type ListJournalsInput struct {
	TenantID uuid.UUID
	Status   *domain.JournalStatus
	Limit    int
	Offset   int
}

// ListJournals returns journals for a tenant.
func (d *Deps) ListJournals(ctx context.Context, in ListJournalsInput) ([]domain.Journal, int, error) {
	if in.TenantID == uuid.Nil {
		return nil, 0, fmt.Errorf("%w: tenant_id", domain.ErrInvalidArgument)
	}
	limit := in.Limit
	if limit <= 0 {
		limit = 50
	}
	return d.Journals.List(ctx, ports.JournalFilter{
		TenantID: in.TenantID,
		Status:   in.Status,
		Limit:    limit,
		Offset:   in.Offset,
	})
}

// GetJournal returns a journal by id.
func (d *Deps) GetJournal(ctx context.Context, tenantID, id uuid.UUID) (domain.Journal, error) {
	if tenantID == uuid.Nil || id == uuid.Nil {
		return domain.Journal{}, fmt.Errorf("%w: tenant_id and id required", domain.ErrInvalidArgument)
	}
	return d.Journals.GetByID(ctx, tenantID, id)
}
