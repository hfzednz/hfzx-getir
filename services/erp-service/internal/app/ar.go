package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/nexora/erp-service/internal/domain"
)

func (d *Deps) CreateARInvoice(ctx context.Context, inv domain.ARInvoice) (domain.ARInvoice, error) {
	if inv.TenantID == uuid.Nil || inv.CompanyID == uuid.Nil || inv.InvoiceNumber == "" || inv.TotalMinor <= 0 {
		return inv, domain.ErrInvalidArgument
	}
	if inv.ID == uuid.Nil {
		inv.ID = d.newID()
	}
	if inv.Currency == "" {
		inv.Currency = "TRY"
	}
	inv.Status = "issued"
	inv.CreatedAt = d.now()
	return inv, d.AR.Save(ctx, inv)
}
