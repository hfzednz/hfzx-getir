package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// Account types for ERP COA (maps to ledger).
const (
	AccountAsset     = "asset"
	AccountLiability = "liability"
	AccountEquity    = "equity"
	AccountRevenue   = "revenue"
	AccountExpense   = "expense"
)

// Money is int64 minor units (never float).
type Money struct {
	AmountMinor int64
	Currency    string
}

func (m Money) Add(o Money) (Money, error) {
	if m.Currency != "" && o.Currency != "" && m.Currency != o.Currency {
		return Money{}, ErrInvalidArgument
	}
	cur := m.Currency
	if cur == "" {
		cur = o.Currency
	}
	return Money{AmountMinor: m.AmountMinor + o.AmountMinor, Currency: cur}, nil
}

// Company is a legal entity book.
type Company struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Code      string
	Name      string
	Country   string
	Currency  string
	Active    bool
	CreatedAt time.Time
}

// FiscalYear groups periods.
type FiscalYear struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	CompanyID uuid.UUID
	Label     string // 2026
	StartDate time.Time
	EndDate   time.Time
	Closed    bool
}

// AccountingPeriod is a month/quarter open|closed.
type AccountingPeriod struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	CompanyID    uuid.UUID
	FiscalYearID uuid.UUID
	Label        string
	StartDate    time.Time
	EndDate      time.Time
	Status       string // open|closed
}

// ChartAccount is ERP COA row.
type ChartAccount struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	CompanyID uuid.UUID
	Code      string
	Name      string
	Type      string
	ParentID  *uuid.UUID
	Active    bool
}

// CostCenter / ProfitCenter.
type CostCenter struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	CompanyID uuid.UUID
	Code      string
	Name      string
}

// JournalEntry is an ERP journal that posts to ledger via port.
type JournalEntry struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	CompanyID      uuid.UUID
	PeriodID       uuid.UUID
	Memo           string
	Currency       string
	Lines          []JournalLine
	Status         string // draft|posted|void
	LedgerRef      string // opaque ledger journal id
	IdempotencyKey string
	CreatedBy      uuid.UUID
	CreatedAt      time.Time
	PostedAt       *time.Time
}

type JournalLine struct {
	AccountCode string
	CostCenter  string
	DebitMinor  int64
	CreditMinor int64
	Memo        string
}

// ValidateBalance ensures debits == credits.
func (j *JournalEntry) ValidateBalance() error {
	var d, c int64
	for _, l := range j.Lines {
		if l.DebitMinor < 0 || l.CreditMinor < 0 {
			return ErrInvalidArgument
		}
		if l.DebitMinor > 0 && l.CreditMinor > 0 {
			return ErrInvalidArgument
		}
		d += l.DebitMinor
		c += l.CreditMinor
	}
	if d == 0 || d != c {
		return ErrUnbalanced
	}
	return nil
}

// Supplier master.
type Supplier struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	CompanyID    uuid.UUID
	Code         string
	Name         string
	TaxID        string
	Country      string
	Currency     string
	RiskScore    float64
	Status       string // pending|active|blocked
	CreatedAt    time.Time
}

// PurchaseRequest → PO workflow.
type PurchaseRequest struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	CompanyID   uuid.UUID
	RequesterID uuid.UUID
	Status      string // draft|submitted|approved|rejected|ordered
	Lines       []ProcurementLine
	Currency    string
	TotalMinor  int64
	CreatedAt   time.Time
}

type ProcurementLine struct {
	SKU         string
	Description string
	Qty         int64
	UnitMinor   int64
	AccountCode string
}

func LinesTotal(lines []ProcurementLine) int64 {
	var t int64
	for _, l := range lines {
		t += l.Qty * l.UnitMinor
	}
	return t
}

// PurchaseOrder.
type PurchaseOrder struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	CompanyID    uuid.UUID
	SupplierID   uuid.UUID
	PRID         *uuid.UUID
	Status       string // open|partial|received|closed|cancelled
	Lines        []ProcurementLine
	Currency     string
	TotalMinor   int64
	CreatedAt    time.Time
}

// GoodsReceipt.
type GoodsReceipt struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	CompanyID  uuid.UUID
	POID       uuid.UUID
	Lines      []ProcurementLine
	ReceivedAt time.Time
	CreatedBy  uuid.UUID
}

// APInvoice supplier bill.
type APInvoice struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	CompanyID     uuid.UUID
	SupplierID    uuid.UUID
	POID          *uuid.UUID
	InvoiceNumber string
	Currency      string
	SubtotalMinor int64
	TaxMinor      int64
	TotalMinor    int64
	Status        string // draft|matched|approved|scheduled|paid|void
	MatchScore    float64
	DueDate       time.Time
	CreatedAt     time.Time
	ApprovedAt    *time.Time
}

// ARInvoice customer bill (corporate AR, not checkout).
type ARInvoice struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	CompanyID     uuid.UUID
	CustomerRef   string
	InvoiceNumber string
	Currency      string
	TotalMinor    int64
	Status        string // draft|issued|paid|void
	DueDate       time.Time
	CreatedAt     time.Time
}

// BankAccount treasury.
type BankAccount struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	CompanyID    uuid.UUID
	Name         string
	IBAN         string
	Currency     string
	BalanceMinor int64
}

// BankTransaction.
type BankTransaction struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	BankAccountID uuid.UUID
	ExternalRef  string
	AmountMinor  int64
	Currency     string
	BookedAt     time.Time
	Reconciled   bool
	Memo         string
}

// Budget.
type Budget struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	CompanyID uuid.UUID
	Label     string
	Period    string // annual|quarterly|monthly
	Year      int
	Status    string // draft|approved|locked
	Currency  string
	Lines     []BudgetLine
	CreatedAt time.Time
	ApprovedAt *time.Time
}

type BudgetLine struct {
	AccountCode string
	CostCenter  string
	AmountMinor int64
	ActualMinor int64
}

func (b *Budget) Variance() []BudgetLine {
	out := make([]BudgetLine, len(b.Lines))
	copy(out, b.Lines)
	return out
}

// FixedAsset.
type FixedAsset struct {
	ID              uuid.UUID
	TenantID        uuid.UUID
	CompanyID       uuid.UUID
	Code            string
	Name            string
	CostMinor       int64
	Currency        string
	UsefulLifeMonths int
	AccumDepMinor   int64
	Status          string // active|disposed
	AcquiredAt      time.Time
}

// StraightLineDepreciation monthly amount.
func StraightLineDepreciation(cost int64, lifeMonths int) int64 {
	if lifeMonths <= 0 {
		return 0
	}
	return cost / int64(lifeMonths)
}

// TaxReturn pack.
type TaxReturn struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	CompanyID    uuid.UUID
	Kind         string // vat|corporate|withholding
	PeriodLabel  string
	Currency     string
	TaxableMinor int64
	TaxMinor     int64
	Status       string // draft|filed
	CreatedAt    time.Time
}

// ExpenseReport.
type ExpenseReport struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	CompanyID   uuid.UUID
	EmployeeID  uuid.UUID
	Currency    string
	TotalMinor  int64
	Status      string // draft|submitted|approved|rejected|paid
	Lines       []ExpenseLine
	CreatedAt   time.Time
}

type ExpenseLine struct {
	Category    string
	AmountMinor int64
	Memo        string
}

// ApprovalRequest workflow instance.
type ApprovalRequest struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	CompanyID  uuid.UUID
	Kind       string // purchase|budget|invoice|expense
	SubjectID  uuid.UUID
	Status     string // pending|approved|rejected
	Steps      []ApprovalStep
	CreatedAt  time.Time
	DecidedAt  *time.Time
}

type ApprovalStep struct {
	Level      int
	ApproverID uuid.UUID
	Status     string // pending|approved|rejected
	Note       string
	DecidedAt  *time.Time
}

// PayrollBatch integration stub.
type PayrollBatch struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	CompanyID   uuid.UUID
	Label       string
	Currency    string
	TotalMinor  int64
	Status      string // draft|exported|acknowledged
	ExternalRef string
	CreatedAt   time.Time
}

// ValidAccountType.
func ValidAccountType(t string) bool {
	switch t {
	case AccountAsset, AccountLiability, AccountEquity, AccountRevenue, AccountExpense:
		return true
	default:
		return false
	}
}

// CanApproveKind.
func ValidApprovalKind(k string) bool {
	switch k {
	case "purchase", "budget", "invoice", "expense":
		return true
	default:
		return false
	}
}

// NormalizeCode uppercases codes.
func NormalizeCode(c string) string {
	return strings.ToUpper(strings.TrimSpace(c))
}

// ThreeWayMatchScore compares PO vs GRN vs invoice totals (0..1).
func ThreeWayMatchScore(poTotal, grnTotal, invTotal int64) float64 {
	if poTotal <= 0 {
		return 0
	}
	diffPOGRN := abs64(poTotal - grnTotal)
	diffPOInv := abs64(poTotal - invTotal)
	pen := float64(diffPOGRN+diffPOInv) / float64(poTotal*2)
	score := 1 - pen
	if score < 0 {
		return 0
	}
	if score > 1 {
		return 1
	}
	return score
}

func abs64(v int64) int64 {
	if v < 0 {
		return -v
	}
	return v
}

// VATAmount computes tax from net (rate in basis points, e.g. 2000 = 20%).
func VATAmount(netMinor int64, rateBps int64) int64 {
	if rateBps < 0 {
		return 0
	}
	return (netMinor * rateBps) / 10000
}
