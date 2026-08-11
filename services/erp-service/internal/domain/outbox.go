package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	OutboxStatusPending   = "pending"
	OutboxStatusPublished = "published"
	OutboxStatusFailed    = "failed"
)

const (
	EventJournalCreated         = "JournalCreated"
	EventInvoiceApproved        = "InvoiceApproved"
	EventBudgetApproved         = "BudgetApproved"
	EventPurchaseOrderCreated   = "PurchaseOrderCreated"
	EventSupplierAdded          = "SupplierAdded"
	EventPaymentScheduled       = "PaymentScheduled"
	EventAssetCreated           = "AssetCreated"
	EventTaxCalculated          = "TaxCalculated"
	EventGoodsReceived          = "GoodsReceived"
	EventExpenseSubmitted       = "ExpenseSubmitted"
)

const TopicERPEvents = "erp.events"

func TopicForEvent(string) string { return TopicERPEvents }

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
