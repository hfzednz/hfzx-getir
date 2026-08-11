package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// LabelKind classifies printable labels.
type LabelKind string

const (
	LabelKindShipping LabelKind = "shipping"
	LabelKindQR       LabelKind = "qr"
	LabelKindBarcode  LabelKind = "barcode"
	LabelKindInternal LabelKind = "internal"
	LabelKindCourier  LabelKind = "courier"
	LabelKindReturn   LabelKind = "return"

	LabelShipping = LabelKindShipping
	LabelQR       = LabelKindQR
	LabelBarcode  = LabelKindBarcode
	LabelInternal = LabelKindInternal
	LabelCourier  = LabelKindCourier
	LabelReturn   = LabelKindReturn
)

func (k LabelKind) Valid() bool {
	switch k {
	case LabelKindShipping, LabelKindQR, LabelKindBarcode, LabelKindInternal, LabelKindCourier, LabelKindReturn:
		return true
	default:
		return false
	}
}

// LabelStatus is print lifecycle.
type LabelStatus string

const (
	LabelStatusDraft   LabelStatus = "draft"
	LabelStatusReady   LabelStatus = "ready"
	LabelStatusPrinted LabelStatus = "printed"
	LabelStatusVoid    LabelStatus = "void"

	LabelDraft   = LabelStatusDraft
	LabelReady   = LabelStatusReady
	LabelPrinted = LabelStatusPrinted
	LabelVoid    = LabelStatusVoid
)

func (s LabelStatus) Valid() bool {
	switch s {
	case LabelStatusDraft, LabelStatusReady, LabelStatusPrinted, LabelStatusVoid:
		return true
	default:
		return false
	}
}

// Label is metadata + print intent for a package/unit.
type Label struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	WarehouseID    uuid.UUID
	FulfillmentID  uuid.UUID
	PackSessionID  uuid.UUID
	DispatchUnitID *uuid.UUID
	Kind           LabelKind
	Status         LabelStatus
	Code           string
	TrackingCode   string
	Barcode        string
	Format         string
	PrintIntent    string
	Payload        map[string]any
	PrinterID      *uuid.UUID
	PrintedAt      *time.Time
	VoidedAt       *time.Time
	Metadata       map[string]any
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Validate checks label invariants.
func (l Label) Validate() error {
	if l.ID == uuid.Nil {
		return fmt.Errorf("%w: label id required", ErrInvalidArgument)
	}
	if l.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if l.TrackingCode == "" && l.Code == "" && l.Barcode == "" {
		return fmt.Errorf("%w: label code/tracking required", ErrInvalidArgument)
	}
	if l.Kind != "" && !l.Kind.Valid() {
		return fmt.Errorf("%w: invalid label kind %q", ErrInvalidArgument, l.Kind)
	}
	if l.Status != "" && !l.Status.Valid() {
		return fmt.Errorf("%w: invalid label status %q", ErrInvalidArgument, l.Status)
	}
	return nil
}

// MarkReady moves draft → ready for print.
func (l *Label) MarkReady() error {
	if l.Status != "" && l.Status != LabelStatusDraft {
		return fmt.Errorf("%w: ready from %s", ErrInvalidTransition, l.Status)
	}
	l.Status = LabelStatusReady
	l.UpdatedAt = time.Now().UTC()
	return nil
}

// MarkPrinted records a successful print.
func (l *Label) MarkPrinted(printerID uuid.UUID) error {
	if l.Status == LabelStatusVoid {
		return fmt.Errorf("%w: print from void", ErrInvalidTransition)
	}
	now := time.Now().UTC()
	l.Status = LabelStatusPrinted
	if printerID != uuid.Nil {
		l.PrinterID = &printerID
	}
	l.PrintedAt = &now
	l.UpdatedAt = now
	return nil
}

// Void cancels a label.
func (l *Label) Void() error {
	if l.Status == LabelStatusVoid {
		return fmt.Errorf("%w: already void", ErrAlreadyTerminal)
	}
	now := time.Now().UTC()
	l.Status = LabelStatusVoid
	l.VoidedAt = &now
	l.UpdatedAt = now
	return nil
}
