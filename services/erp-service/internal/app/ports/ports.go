package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/erp-service/internal/domain"
)

type Clock interface{ Now() time.Time }
type IDGen interface{ New() uuid.UUID }

type CompanyRepo interface {
	Save(ctx context.Context, c domain.Company) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Company, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.Company, error)
}

type PeriodRepo interface {
	SaveYear(ctx context.Context, y domain.FiscalYear) error
	SavePeriod(ctx context.Context, p domain.AccountingPeriod) error
	GetPeriod(ctx context.Context, tenantID, id uuid.UUID) (domain.AccountingPeriod, error)
	ListPeriods(ctx context.Context, tenantID, companyID uuid.UUID) ([]domain.AccountingPeriod, error)
}

type AccountRepo interface {
	Save(ctx context.Context, a domain.ChartAccount) error
	GetByCode(ctx context.Context, tenantID, companyID uuid.UUID, code string) (domain.ChartAccount, error)
	List(ctx context.Context, tenantID, companyID uuid.UUID) ([]domain.ChartAccount, error)
}

type JournalRepo interface {
	Save(ctx context.Context, j domain.JournalEntry) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.JournalEntry, error)
	GetByIdempotency(ctx context.Context, tenantID uuid.UUID, key string) (domain.JournalEntry, bool, error)
}

type SupplierRepo interface {
	Save(ctx context.Context, s domain.Supplier) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Supplier, error)
	List(ctx context.Context, tenantID, companyID uuid.UUID) ([]domain.Supplier, error)
}

type ProcurementRepo interface {
	SavePR(ctx context.Context, p domain.PurchaseRequest) error
	GetPR(ctx context.Context, tenantID, id uuid.UUID) (domain.PurchaseRequest, error)
	SavePO(ctx context.Context, p domain.PurchaseOrder) error
	GetPO(ctx context.Context, tenantID, id uuid.UUID) (domain.PurchaseOrder, error)
	SaveGRN(ctx context.Context, g domain.GoodsReceipt) error
	ListGRNByPO(ctx context.Context, tenantID, poID uuid.UUID) ([]domain.GoodsReceipt, error)
}

type APRepo interface {
	Save(ctx context.Context, inv domain.APInvoice) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.APInvoice, error)
	List(ctx context.Context, tenantID, companyID uuid.UUID, status string) ([]domain.APInvoice, error)
}

type ARRepo interface {
	Save(ctx context.Context, inv domain.ARInvoice) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.ARInvoice, error)
}

type TreasuryRepo interface {
	SaveBank(ctx context.Context, b domain.BankAccount) error
	GetBank(ctx context.Context, tenantID, id uuid.UUID) (domain.BankAccount, error)
	ListBanks(ctx context.Context, tenantID, companyID uuid.UUID) ([]domain.BankAccount, error)
	SaveTxn(ctx context.Context, t domain.BankTransaction) error
	ListTxns(ctx context.Context, tenantID, bankID uuid.UUID) ([]domain.BankTransaction, error)
}

type BudgetRepo interface {
	Save(ctx context.Context, b domain.Budget) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Budget, error)
	List(ctx context.Context, tenantID, companyID uuid.UUID) ([]domain.Budget, error)
}

type AssetRepo interface {
	Save(ctx context.Context, a domain.FixedAsset) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.FixedAsset, error)
	List(ctx context.Context, tenantID, companyID uuid.UUID) ([]domain.FixedAsset, error)
}

type TaxRepo interface {
	Save(ctx context.Context, t domain.TaxReturn) error
	List(ctx context.Context, tenantID, companyID uuid.UUID) ([]domain.TaxReturn, error)
}

type ExpenseRepo interface {
	Save(ctx context.Context, e domain.ExpenseReport) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.ExpenseReport, error)
}

type ApprovalRepo interface {
	Save(ctx context.Context, a domain.ApprovalRequest) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.ApprovalRequest, error)
	ListPending(ctx context.Context, tenantID uuid.UUID) ([]domain.ApprovalRequest, error)
}

type PayrollRepo interface {
	Save(ctx context.Context, p domain.PayrollBatch) error
	List(ctx context.Context, tenantID, companyID uuid.UUID) ([]domain.PayrollBatch, error)
}

type OutboxRepository interface {
	Enqueue(ctx context.Context, m domain.OutboxMessage) error
	ListPending(ctx context.Context, limit int) ([]domain.OutboxMessage, error)
	Update(ctx context.Context, m domain.OutboxMessage) error
}

type EventPublisher interface {
	Publish(ctx context.Context, topic, key string, payload map[string]any) error
}

// LedgerClient posts balanced journals to finance-ledger-service.
type LedgerClient interface {
	PostJournal(ctx context.Context, tenantID, companyID uuid.UUID, memo string, currency string, lines []domain.JournalLine, idemKey string) (ledgerRef string, err error)
}

// InventoryClient receives goods into inventory (no stock SoT here).
type InventoryClient interface {
	Receive(ctx context.Context, tenantID uuid.UUID, sku string, qty int64, warehouseRef string) error
}

// SettlementClient schedules supplier payouts (execution elsewhere).
type SettlementClient interface {
	ScheduleSupplierPayment(ctx context.Context, tenantID, supplierID uuid.UUID, amountMinor int64, currency string, invoiceRef string) (batchRef string, err error)
}

// AIClient optional cashflow/risk hints.
type AIClient interface {
	PredictCashflow(ctx context.Context, tenantID, companyID uuid.UUID, horizonDays int) (map[string]float64, error)
	SupplierRisk(ctx context.Context, tenantID uuid.UUID, features map[string]float64) (float64, error)
}
