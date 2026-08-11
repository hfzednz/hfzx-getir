package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	OutboxStatusPending   = "pending"
	OutboxStatusPublished = "published"
	OutboxStatusFailed    = "failed"
	TopicSupplierEvents   = "supplier.events"
)

const (
	EventSupplierCreated      = "SupplierCreated"
	EventSupplierApproved     = "SupplierApproved"
	EventPurchaseOrderCreated = "PurchaseOrderCreated"
	EventShipmentReceived     = "ShipmentReceived"
	EventInvoiceMatched       = "InvoiceMatched"
	EventContractRenewed      = "ContractRenewed"
	EventSellerOnboarded      = "SellerOnboarded"
	EventSupplierRated        = "SupplierRated"
)

func TopicForEvent(string) string { return TopicSupplierEvents }

type OutboxMessage struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	AggregateID uuid.UUID
	Topic       string
	Key         string
	Payload     map[string]any
	Status      string
	Attempts    int
	LastError   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	PublishedAt *time.Time
}
