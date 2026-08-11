package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/supplier-service/internal/domain"
)

type Clock interface{ Now() time.Time }
type IDGen interface{ New() uuid.UUID }

type OutboxRepository interface {
	Enqueue(ctx context.Context, m domain.OutboxMessage) error
	ListPending(ctx context.Context, limit int) ([]domain.OutboxMessage, error)
	Update(ctx context.Context, m domain.OutboxMessage) error
}

type EventPublisher interface {
	Publish(ctx context.Context, topic, key string, payload map[string]any) error
}

type SupplierRepo interface {
	Save(ctx context.Context, s domain.Supplier) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Supplier, error)
	GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (domain.Supplier, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.Supplier, error)
}

type DocumentRepo interface {
	Save(ctx context.Context, d domain.SupplierDocument) error
	ListBySupplier(ctx context.Context, tenantID, supplierID uuid.UUID) ([]domain.SupplierDocument, error)
}

type CertRepo interface {
	Save(ctx context.Context, c domain.Certification) error
	ListBySupplier(ctx context.Context, tenantID, supplierID uuid.UUID) ([]domain.Certification, error)
}

type ContractRepo interface {
	Save(ctx context.Context, c domain.Contract) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Contract, error)
	ListBySupplier(ctx context.Context, tenantID, supplierID uuid.UUID) ([]domain.Contract, error)
}

type RFQRepo interface {
	Save(ctx context.Context, r domain.RFQ) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.RFQ, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.RFQ, error)
}

type QuotationRepo interface {
	Save(ctx context.Context, q domain.Quotation) error
	ListByRFQ(ctx context.Context, tenantID, rfqID uuid.UUID) ([]domain.Quotation, error)
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Quotation, error)
}

type PORepo interface {
	Save(ctx context.Context, po domain.SourcingPurchaseOrder) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.SourcingPurchaseOrder, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.SourcingPurchaseOrder, error)
	ListBySupplier(ctx context.Context, tenantID, supplierID uuid.UUID) ([]domain.SourcingPurchaseOrder, error)
}

type ShipmentRepo interface {
	Save(ctx context.Context, s domain.InboundShipment) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.InboundShipment, error)
	ListByPO(ctx context.Context, tenantID, poID uuid.UUID) ([]domain.InboundShipment, error)
}

type InvoiceMatchRepo interface {
	Save(ctx context.Context, m domain.InvoiceMatchSignal) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.InvoiceMatchSignal, error)
}

type SellerRepo interface {
	Save(ctx context.Context, s domain.MarketplaceSeller) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.MarketplaceSeller, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.MarketplaceSeller, error)
}

type ListingRepo interface {
	Save(ctx context.Context, l domain.ListingRef) error
	ListBySeller(ctx context.Context, tenantID, sellerID uuid.UUID) ([]domain.ListingRef, error)
}

type SubmissionRepo interface {
	Save(ctx context.Context, s domain.CatalogSubmission) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.CatalogSubmission, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.CatalogSubmission, error)
}

type EDIRepo interface {
	Save(ctx context.Context, d domain.EDIDocument) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.EDIDocument, error)
}

type ScorecardRepo interface {
	Save(ctx context.Context, s domain.Scorecard) error
	ListBySupplier(ctx context.Context, tenantID, supplierID uuid.UUID) ([]domain.Scorecard, error)
}

type MessageRepo interface {
	SaveThread(ctx context.Context, t domain.MessageThread) error
	SaveMessage(ctx context.Context, m domain.Message) error
	ListThreads(ctx context.Context, tenantID, supplierID uuid.UUID) ([]domain.MessageThread, error)
	ListMessages(ctx context.Context, tenantID, threadID uuid.UUID) ([]domain.Message, error)
}

type ChangeRepo interface {
	Save(ctx context.Context, c domain.ChangeRequest) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.ChangeRequest, error)
}

// External ports — no ownership of remotes.
type ERPClient interface {
	SyncSupplier(ctx context.Context, tenantID uuid.UUID, supplier domain.Supplier) (erpSupplierID string, err error)
	EnsurePO(ctx context.Context, tenantID uuid.UUID, po domain.SourcingPurchaseOrder) (erpPOID string, err error)
}

type CatalogClient interface {
	SubmitProduct(ctx context.Context, tenantID uuid.UUID, sub domain.CatalogSubmission) error
}

type InventoryClient interface {
	AnnounceASN(ctx context.Context, tenantID uuid.UUID, ship domain.InboundShipment) error
}

type SettlementClient interface {
	ScheduleSellerPayout(ctx context.Context, tenantID, sellerID uuid.UUID, amountMinor int64, currency string) error
}

type AIClient interface {
	RecommendSuppliers(ctx context.Context, tenantID uuid.UUID, sku string, limit int) ([]uuid.UUID, error)
	PredictRisk(ctx context.Context, supplier domain.Supplier) (float64, error)
	RankSuppliers(ctx context.Context, cards []domain.Scorecard) ([]uuid.UUID, error)
}

type MetricsClient interface {
	Record(ctx context.Context, name string, tags map[string]string, value float64) error
}
