package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/finance-ledger-service/internal/domain"
)

// InvoiceLineInput is a line for CreateInvoice.
type InvoiceLineInput struct {
	Description string
	Qty         int64
	UnitMinor   int64
	TaxCode     string
}

// CreateInvoiceInput creates a draft/issued invoice with tax calculation.
type CreateInvoiceInput struct {
	TenantID        uuid.UUID
	Currency        string
	CounterpartyRef string
	ExternalRef     string
	IdempotencyKey  string
	Issue           bool
	Lines           []InvoiceLineInput
	DefaultTaxCode  string
}

// CreateInvoice builds an invoice, applying tax rules when tax codes are set.
func (d *Deps) CreateInvoice(ctx context.Context, in CreateInvoiceInput) (domain.Invoice, error) {
	if in.TenantID == uuid.Nil {
		return domain.Invoice{}, fmt.Errorf("%w: tenant_id", domain.ErrInvalidArgument)
	}
	currency := strings.ToUpper(strings.TrimSpace(in.Currency))
	if _, err := domain.NewMoney(0, currency); err != nil {
		return domain.Invoice{}, err
	}
	if in.IdempotencyKey != "" {
		if existing, err := d.Invoices.GetByIdempotencyKey(ctx, in.TenantID, in.IdempotencyKey); err == nil {
			return existing, nil
		} else if !errors.Is(err, domain.ErrNotFound) {
			return domain.Invoice{}, err
		}
	}
	if len(in.Lines) == 0 {
		return domain.Invoice{}, fmt.Errorf("%w: lines required", domain.ErrInvalidArgument)
	}

	now := d.now()
	lines := make([]domain.InvoiceLine, 0, len(in.Lines))
	var subtotal, taxTotal int64
	for _, li := range in.Lines {
		if li.Qty <= 0 {
			return domain.Invoice{}, fmt.Errorf("%w: qty", domain.ErrInvalidArgument)
		}
		if li.UnitMinor < 0 {
			return domain.Invoice{}, fmt.Errorf("%w: unit", domain.ErrNegativeMoney)
		}
		base := li.Qty * li.UnitMinor
		taxCode := strings.TrimSpace(li.TaxCode)
		if taxCode == "" {
			taxCode = strings.TrimSpace(in.DefaultTaxCode)
		}
		var taxMinor int64
		if taxCode != "" {
			tr, err := d.TaxCalculate(ctx, TaxCalculateInput{
				TenantID:  in.TenantID,
				BaseMinor: base,
				Currency:  currency,
				TaxCode:   taxCode,
			})
			if err != nil {
				return domain.Invoice{}, err
			}
			taxMinor = tr.TaxMinor
		}
		line := domain.InvoiceLine{
			ID:          d.newID(),
			Description: strings.TrimSpace(li.Description),
			Qty:         li.Qty,
			UnitMinor:   li.UnitMinor,
			TaxMinor:    taxMinor,
			TotalMinor:  base + taxMinor,
			TaxCode:     taxCode,
		}
		if err := line.Validate(); err != nil {
			return domain.Invoice{}, err
		}
		lines = append(lines, line)
		subtotal += base
		taxTotal += taxMinor
	}

	status := domain.InvoiceStatusDraft
	var issuedAt *time.Time
	if in.Issue {
		status = domain.InvoiceStatusIssued
		issuedAt = &now
	}
	inv := domain.Invoice{
		ID:              d.newID(),
		TenantID:        in.TenantID,
		Status:          status,
		Currency:        currency,
		CounterpartyRef: in.CounterpartyRef,
		ExternalRef:     in.ExternalRef,
		IdempotencyKey:  in.IdempotencyKey,
		SubtotalMinor:   subtotal,
		TaxMinor:        taxTotal,
		TotalMinor:      subtotal + taxTotal,
		Lines:           lines,
		IssuedAt:        issuedAt,
		CreatedAt:       now,
		UpdatedAt:       now,
		Version:         1,
	}
	if err := inv.Validate(); err != nil {
		return domain.Invoice{}, err
	}
	if err := d.Invoices.Create(ctx, inv); err != nil {
		return domain.Invoice{}, err
	}
	_ = d.appendEvent(ctx, inv.ID, inv.TenantID, domain.EventInvoiceGenerated, map[string]any{
		"status": string(inv.Status), "totalMinor": inv.TotalMinor, "currency": inv.Currency,
	})
	return inv, nil
}

// IssueCreditNoteInput issues a credit note against an invoice.
type IssueCreditNoteInput struct {
	TenantID       uuid.UUID
	InvoiceID      uuid.UUID
	AmountMinor    int64
	Reason         string
	IdempotencyKey string
}

// IssueCreditNote creates a credit note and marks invoice credited when fully covered.
func (d *Deps) IssueCreditNote(ctx context.Context, in IssueCreditNoteInput) (domain.CreditNote, error) {
	if in.TenantID == uuid.Nil || in.InvoiceID == uuid.Nil {
		return domain.CreditNote{}, fmt.Errorf("%w: tenant_id and invoice_id required", domain.ErrInvalidArgument)
	}
	if in.AmountMinor <= 0 {
		return domain.CreditNote{}, fmt.Errorf("%w: amount", domain.ErrInvalidArgument)
	}
	if in.IdempotencyKey != "" {
		if existing, err := d.Invoices.GetCreditNoteByIdempotencyKey(ctx, in.TenantID, in.IdempotencyKey); err == nil {
			return existing, nil
		} else if !errors.Is(err, domain.ErrNotFound) {
			return domain.CreditNote{}, err
		}
	}
	inv, err := d.Invoices.GetByID(ctx, in.TenantID, in.InvoiceID)
	if err != nil {
		return domain.CreditNote{}, err
	}
	switch inv.Status {
	case domain.InvoiceStatusIssued, domain.InvoiceStatusPaid, domain.InvoiceStatusCredited:
		// ok
	default:
		return domain.CreditNote{}, fmt.Errorf("%w: cannot credit invoice in status %s", domain.ErrInvoiceImmutable, inv.Status)
	}
	if in.AmountMinor > inv.TotalMinor {
		return domain.CreditNote{}, fmt.Errorf("%w: credit exceeds invoice total", domain.ErrInvalidArgument)
	}
	now := d.now()
	cn := domain.CreditNote{
		ID:             d.newID(),
		TenantID:       in.TenantID,
		InvoiceID:      inv.ID,
		Currency:       inv.Currency,
		AmountMinor:    in.AmountMinor,
		Reason:         in.Reason,
		IdempotencyKey: in.IdempotencyKey,
		CreatedAt:      now,
	}
	if err := cn.Validate(); err != nil {
		return domain.CreditNote{}, err
	}
	if err := d.Invoices.CreateCreditNote(ctx, cn); err != nil {
		return domain.CreditNote{}, err
	}
	if in.AmountMinor == inv.TotalMinor {
		inv.Status = domain.InvoiceStatusCredited
		inv.UpdatedAt = now
		inv.Version++
		if err := d.Invoices.Update(ctx, inv); err != nil {
			return domain.CreditNote{}, err
		}
	}
	_ = d.appendEvent(ctx, cn.ID, cn.TenantID, domain.EventCreditNoteIssued, map[string]any{
		"invoiceId": inv.ID.String(), "amountMinor": cn.AmountMinor,
	})
	return cn, nil
}
