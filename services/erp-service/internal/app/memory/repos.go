package memory

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/nexora/erp-service/internal/domain"
)

type Store struct {
	mu        sync.RWMutex
	Companies []domain.Company
	Years     []domain.FiscalYear
	Periods   []domain.AccountingPeriod
	Accounts  []domain.ChartAccount
	Journals  []domain.JournalEntry
	JournIdem map[string]uuid.UUID
	Suppliers []domain.Supplier
	PRs       []domain.PurchaseRequest
	POs       []domain.PurchaseOrder
	GRNs      []domain.GoodsReceipt
	AP        []domain.APInvoice
	AR        []domain.ARInvoice
	Banks     []domain.BankAccount
	Txns      []domain.BankTransaction
	Budgets   []domain.Budget
	Assets    []domain.FixedAsset
	Tax       []domain.TaxReturn
	Expenses  []domain.ExpenseReport
	Approvals []domain.ApprovalRequest
	Payroll   []domain.PayrollBatch
	Outbox    []domain.OutboxMessage
}

func NewStore() *Store {
	return &Store{JournIdem: make(map[string]uuid.UUID)}
}

type Repos struct {
	Companies   *CompanyRepo
	Periods     *PeriodRepo
	Accounts    *AccountRepo
	Journals    *JournalRepo
	Suppliers   *SupplierRepo
	Procurement *ProcurementRepo
	AP          *APRepo
	AR          *ARRepo
	Treasury    *TreasuryRepo
	Budgets     *BudgetRepo
	Assets      *AssetRepo
	Tax         *TaxRepo
	Expenses    *ExpenseRepo
	Approvals   *ApprovalRepo
	Payroll     *PayrollRepo
	Outbox      *OutboxRepo
	Ledger      *MockLedger
	Inventory   *MockInventory
	Settlement  *MockSettlement
	AI          *MockAI
}

func NewRepos(s *Store) *Repos {
	return &Repos{
		Companies: &CompanyRepo{s: s}, Periods: &PeriodRepo{s: s}, Accounts: &AccountRepo{s: s},
		Journals: &JournalRepo{s: s}, Suppliers: &SupplierRepo{s: s}, Procurement: &ProcurementRepo{s: s},
		AP: &APRepo{s: s}, AR: &ARRepo{s: s}, Treasury: &TreasuryRepo{s: s}, Budgets: &BudgetRepo{s: s},
		Assets: &AssetRepo{s: s}, Tax: &TaxRepo{s: s}, Expenses: &ExpenseRepo{s: s},
		Approvals: &ApprovalRepo{s: s}, Payroll: &PayrollRepo{s: s}, Outbox: &OutboxRepo{s: s},
		Ledger: &MockLedger{}, Inventory: &MockInventory{}, Settlement: &MockSettlement{}, AI: &MockAI{},
	}
}

type CompanyRepo struct{ s *Store }

func (r *CompanyRepo) Save(_ context.Context, c domain.Company) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Companies {
		if r.s.Companies[i].ID == c.ID {
			r.s.Companies[i] = c
			return nil
		}
	}
	r.s.Companies = append(r.s.Companies, c)
	return nil
}

func (r *CompanyRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.Company, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, c := range r.s.Companies {
		if c.TenantID == tenantID && c.ID == id {
			return c, nil
		}
	}
	return domain.Company{}, domain.ErrNotFound
}

func (r *CompanyRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.Company, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.Company{}
	for _, c := range r.s.Companies {
		if c.TenantID == tenantID {
			out = append(out, c)
		}
	}
	return out, nil
}

type PeriodRepo struct{ s *Store }

func (r *PeriodRepo) SaveYear(_ context.Context, y domain.FiscalYear) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Years {
		if r.s.Years[i].ID == y.ID {
			r.s.Years[i] = y
			return nil
		}
	}
	r.s.Years = append(r.s.Years, y)
	return nil
}

func (r *PeriodRepo) SavePeriod(_ context.Context, p domain.AccountingPeriod) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Periods {
		if r.s.Periods[i].ID == p.ID {
			r.s.Periods[i] = p
			return nil
		}
	}
	r.s.Periods = append(r.s.Periods, p)
	return nil
}

func (r *PeriodRepo) GetPeriod(_ context.Context, tenantID, id uuid.UUID) (domain.AccountingPeriod, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, p := range r.s.Periods {
		if p.TenantID == tenantID && p.ID == id {
			return p, nil
		}
	}
	return domain.AccountingPeriod{}, domain.ErrNotFound
}

func (r *PeriodRepo) ListPeriods(_ context.Context, tenantID, companyID uuid.UUID) ([]domain.AccountingPeriod, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.AccountingPeriod{}
	for _, p := range r.s.Periods {
		if p.TenantID == tenantID && p.CompanyID == companyID {
			out = append(out, p)
		}
	}
	return out, nil
}

type AccountRepo struct{ s *Store }

func (r *AccountRepo) Save(_ context.Context, a domain.ChartAccount) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Accounts {
		if r.s.Accounts[i].ID == a.ID {
			r.s.Accounts[i] = a
			return nil
		}
	}
	r.s.Accounts = append(r.s.Accounts, a)
	return nil
}

func (r *AccountRepo) GetByCode(_ context.Context, tenantID, companyID uuid.UUID, code string) (domain.ChartAccount, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	code = domain.NormalizeCode(code)
	for _, a := range r.s.Accounts {
		if a.TenantID == tenantID && a.CompanyID == companyID && a.Code == code {
			return a, nil
		}
	}
	return domain.ChartAccount{}, domain.ErrNotFound
}

func (r *AccountRepo) List(_ context.Context, tenantID, companyID uuid.UUID) ([]domain.ChartAccount, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.ChartAccount{}
	for _, a := range r.s.Accounts {
		if a.TenantID == tenantID && a.CompanyID == companyID {
			out = append(out, a)
		}
	}
	return out, nil
}

type JournalRepo struct{ s *Store }

func (r *JournalRepo) Save(_ context.Context, j domain.JournalEntry) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Journals {
		if r.s.Journals[i].ID == j.ID {
			r.s.Journals[i] = j
			if j.IdempotencyKey != "" {
				r.s.JournIdem[j.TenantID.String()+"|"+j.IdempotencyKey] = j.ID
			}
			return nil
		}
	}
	r.s.Journals = append(r.s.Journals, j)
	if j.IdempotencyKey != "" {
		r.s.JournIdem[j.TenantID.String()+"|"+j.IdempotencyKey] = j.ID
	}
	return nil
}

func (r *JournalRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.JournalEntry, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, j := range r.s.Journals {
		if j.TenantID == tenantID && j.ID == id {
			return j, nil
		}
	}
	return domain.JournalEntry{}, domain.ErrNotFound
}

func (r *JournalRepo) GetByIdempotency(_ context.Context, tenantID uuid.UUID, key string) (domain.JournalEntry, bool, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	id, ok := r.s.JournIdem[tenantID.String()+"|"+key]
	if !ok {
		return domain.JournalEntry{}, false, nil
	}
	for _, j := range r.s.Journals {
		if j.ID == id {
			return j, true, nil
		}
	}
	return domain.JournalEntry{}, false, nil
}

type SupplierRepo struct{ s *Store }

func (r *SupplierRepo) Save(_ context.Context, s domain.Supplier) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Suppliers {
		if r.s.Suppliers[i].ID == s.ID {
			r.s.Suppliers[i] = s
			return nil
		}
	}
	r.s.Suppliers = append(r.s.Suppliers, s)
	return nil
}

func (r *SupplierRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.Supplier, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, s := range r.s.Suppliers {
		if s.TenantID == tenantID && s.ID == id {
			return s, nil
		}
	}
	return domain.Supplier{}, domain.ErrNotFound
}

func (r *SupplierRepo) List(_ context.Context, tenantID, companyID uuid.UUID) ([]domain.Supplier, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.Supplier{}
	for _, s := range r.s.Suppliers {
		if s.TenantID == tenantID && s.CompanyID == companyID {
			out = append(out, s)
		}
	}
	return out, nil
}

type ProcurementRepo struct{ s *Store }

func (r *ProcurementRepo) SavePR(_ context.Context, p domain.PurchaseRequest) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.PRs {
		if r.s.PRs[i].ID == p.ID {
			r.s.PRs[i] = p
			return nil
		}
	}
	r.s.PRs = append(r.s.PRs, p)
	return nil
}

func (r *ProcurementRepo) GetPR(_ context.Context, tenantID, id uuid.UUID) (domain.PurchaseRequest, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, p := range r.s.PRs {
		if p.TenantID == tenantID && p.ID == id {
			return p, nil
		}
	}
	return domain.PurchaseRequest{}, domain.ErrNotFound
}

func (r *ProcurementRepo) SavePO(_ context.Context, p domain.PurchaseOrder) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.POs {
		if r.s.POs[i].ID == p.ID {
			r.s.POs[i] = p
			return nil
		}
	}
	r.s.POs = append(r.s.POs, p)
	return nil
}

func (r *ProcurementRepo) GetPO(_ context.Context, tenantID, id uuid.UUID) (domain.PurchaseOrder, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, p := range r.s.POs {
		if p.TenantID == tenantID && p.ID == id {
			return p, nil
		}
	}
	return domain.PurchaseOrder{}, domain.ErrNotFound
}

func (r *ProcurementRepo) SaveGRN(_ context.Context, g domain.GoodsReceipt) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.GRNs {
		if r.s.GRNs[i].ID == g.ID {
			r.s.GRNs[i] = g
			return nil
		}
	}
	r.s.GRNs = append(r.s.GRNs, g)
	return nil
}

func (r *ProcurementRepo) ListGRNByPO(_ context.Context, tenantID, poID uuid.UUID) ([]domain.GoodsReceipt, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.GoodsReceipt{}
	for _, g := range r.s.GRNs {
		if g.TenantID == tenantID && g.POID == poID {
			out = append(out, g)
		}
	}
	return out, nil
}

type APRepo struct{ s *Store }

func (r *APRepo) Save(_ context.Context, inv domain.APInvoice) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.AP {
		if r.s.AP[i].ID == inv.ID {
			r.s.AP[i] = inv
			return nil
		}
	}
	r.s.AP = append(r.s.AP, inv)
	return nil
}

func (r *APRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.APInvoice, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, inv := range r.s.AP {
		if inv.TenantID == tenantID && inv.ID == id {
			return inv, nil
		}
	}
	return domain.APInvoice{}, domain.ErrNotFound
}

func (r *APRepo) List(_ context.Context, tenantID, companyID uuid.UUID, status string) ([]domain.APInvoice, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.APInvoice{}
	for _, inv := range r.s.AP {
		if inv.TenantID == tenantID && inv.CompanyID == companyID && (status == "" || inv.Status == status) {
			out = append(out, inv)
		}
	}
	return out, nil
}

type ARRepo struct{ s *Store }

func (r *ARRepo) Save(_ context.Context, inv domain.ARInvoice) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.AR {
		if r.s.AR[i].ID == inv.ID {
			r.s.AR[i] = inv
			return nil
		}
	}
	r.s.AR = append(r.s.AR, inv)
	return nil
}

func (r *ARRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.ARInvoice, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, inv := range r.s.AR {
		if inv.TenantID == tenantID && inv.ID == id {
			return inv, nil
		}
	}
	return domain.ARInvoice{}, domain.ErrNotFound
}

type TreasuryRepo struct{ s *Store }

func (r *TreasuryRepo) SaveBank(_ context.Context, b domain.BankAccount) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Banks {
		if r.s.Banks[i].ID == b.ID {
			r.s.Banks[i] = b
			return nil
		}
	}
	r.s.Banks = append(r.s.Banks, b)
	return nil
}

func (r *TreasuryRepo) GetBank(_ context.Context, tenantID, id uuid.UUID) (domain.BankAccount, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, b := range r.s.Banks {
		if b.TenantID == tenantID && b.ID == id {
			return b, nil
		}
	}
	return domain.BankAccount{}, domain.ErrNotFound
}

func (r *TreasuryRepo) ListBanks(_ context.Context, tenantID, companyID uuid.UUID) ([]domain.BankAccount, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.BankAccount{}
	for _, b := range r.s.Banks {
		if b.TenantID == tenantID && b.CompanyID == companyID {
			out = append(out, b)
		}
	}
	return out, nil
}

func (r *TreasuryRepo) SaveTxn(_ context.Context, t domain.BankTransaction) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Txns {
		if r.s.Txns[i].ID == t.ID {
			r.s.Txns[i] = t
			return nil
		}
	}
	r.s.Txns = append(r.s.Txns, t)
	return nil
}

func (r *TreasuryRepo) ListTxns(_ context.Context, tenantID, bankID uuid.UUID) ([]domain.BankTransaction, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.BankTransaction{}
	for _, t := range r.s.Txns {
		if t.TenantID == tenantID && t.BankAccountID == bankID {
			out = append(out, t)
		}
	}
	return out, nil
}

type BudgetRepo struct{ s *Store }

func (r *BudgetRepo) Save(_ context.Context, b domain.Budget) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Budgets {
		if r.s.Budgets[i].ID == b.ID {
			r.s.Budgets[i] = b
			return nil
		}
	}
	r.s.Budgets = append(r.s.Budgets, b)
	return nil
}

func (r *BudgetRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.Budget, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, b := range r.s.Budgets {
		if b.TenantID == tenantID && b.ID == id {
			return b, nil
		}
	}
	return domain.Budget{}, domain.ErrNotFound
}

func (r *BudgetRepo) List(_ context.Context, tenantID, companyID uuid.UUID) ([]domain.Budget, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.Budget{}
	for _, b := range r.s.Budgets {
		if b.TenantID == tenantID && b.CompanyID == companyID {
			out = append(out, b)
		}
	}
	return out, nil
}

type AssetRepo struct{ s *Store }

func (r *AssetRepo) Save(_ context.Context, a domain.FixedAsset) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Assets {
		if r.s.Assets[i].ID == a.ID {
			r.s.Assets[i] = a
			return nil
		}
	}
	r.s.Assets = append(r.s.Assets, a)
	return nil
}

func (r *AssetRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.FixedAsset, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, a := range r.s.Assets {
		if a.TenantID == tenantID && a.ID == id {
			return a, nil
		}
	}
	return domain.FixedAsset{}, domain.ErrNotFound
}

func (r *AssetRepo) List(_ context.Context, tenantID, companyID uuid.UUID) ([]domain.FixedAsset, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.FixedAsset{}
	for _, a := range r.s.Assets {
		if a.TenantID == tenantID && a.CompanyID == companyID {
			out = append(out, a)
		}
	}
	return out, nil
}

type TaxRepo struct{ s *Store }

func (r *TaxRepo) Save(_ context.Context, t domain.TaxReturn) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Tax {
		if r.s.Tax[i].ID == t.ID {
			r.s.Tax[i] = t
			return nil
		}
	}
	r.s.Tax = append(r.s.Tax, t)
	return nil
}

func (r *TaxRepo) List(_ context.Context, tenantID, companyID uuid.UUID) ([]domain.TaxReturn, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.TaxReturn{}
	for _, t := range r.s.Tax {
		if t.TenantID == tenantID && t.CompanyID == companyID {
			out = append(out, t)
		}
	}
	return out, nil
}

type ExpenseRepo struct{ s *Store }

func (r *ExpenseRepo) Save(_ context.Context, e domain.ExpenseReport) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Expenses {
		if r.s.Expenses[i].ID == e.ID {
			r.s.Expenses[i] = e
			return nil
		}
	}
	r.s.Expenses = append(r.s.Expenses, e)
	return nil
}

func (r *ExpenseRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.ExpenseReport, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, e := range r.s.Expenses {
		if e.TenantID == tenantID && e.ID == id {
			return e, nil
		}
	}
	return domain.ExpenseReport{}, domain.ErrNotFound
}

type ApprovalRepo struct{ s *Store }

func (r *ApprovalRepo) Save(_ context.Context, a domain.ApprovalRequest) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Approvals {
		if r.s.Approvals[i].ID == a.ID {
			r.s.Approvals[i] = a
			return nil
		}
	}
	r.s.Approvals = append(r.s.Approvals, a)
	return nil
}

func (r *ApprovalRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.ApprovalRequest, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, a := range r.s.Approvals {
		if a.TenantID == tenantID && a.ID == id {
			return a, nil
		}
	}
	return domain.ApprovalRequest{}, domain.ErrNotFound
}

func (r *ApprovalRepo) ListPending(_ context.Context, tenantID uuid.UUID) ([]domain.ApprovalRequest, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.ApprovalRequest{}
	for _, a := range r.s.Approvals {
		if a.TenantID == tenantID && a.Status == "pending" {
			out = append(out, a)
		}
	}
	return out, nil
}

type PayrollRepo struct{ s *Store }

func (r *PayrollRepo) Save(_ context.Context, p domain.PayrollBatch) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Payroll {
		if r.s.Payroll[i].ID == p.ID {
			r.s.Payroll[i] = p
			return nil
		}
	}
	r.s.Payroll = append(r.s.Payroll, p)
	return nil
}

func (r *PayrollRepo) List(_ context.Context, tenantID, companyID uuid.UUID) ([]domain.PayrollBatch, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.PayrollBatch{}
	for _, p := range r.s.Payroll {
		if p.TenantID == tenantID && p.CompanyID == companyID {
			out = append(out, p)
		}
	}
	return out, nil
}

type OutboxRepo struct{ s *Store }

func (r *OutboxRepo) Enqueue(_ context.Context, m domain.OutboxMessage) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Outbox = append(r.s.Outbox, m)
	return nil
}

func (r *OutboxRepo) ListPending(_ context.Context, limit int) ([]domain.OutboxMessage, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.OutboxMessage{}
	for _, m := range r.s.Outbox {
		if m.Status == domain.OutboxStatusPending {
			out = append(out, m)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (r *OutboxRepo) Update(_ context.Context, m domain.OutboxMessage) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Outbox {
		if r.s.Outbox[i].ID == m.ID {
			r.s.Outbox[i] = m
			return nil
		}
	}
	return domain.ErrNotFound
}

// Mock external ports.

type MockLedger struct{ N int }

func (m *MockLedger) PostJournal(_ context.Context, _, _ uuid.UUID, _ string, _ string, _ []domain.JournalLine, _ string) (string, error) {
	m.N++
	return "ledger-ref-" + uuid.NewString()[:8], nil
}

type MockInventory struct{ Received int64 }

func (m *MockInventory) Receive(_ context.Context, _ uuid.UUID, _ string, qty int64, _ string) error {
	m.Received += qty
	return nil
}

type MockSettlement struct{ N int }

func (m *MockSettlement) ScheduleSupplierPayment(_ context.Context, _, _ uuid.UUID, _ int64, _, _ string) (string, error) {
	m.N++
	return "settle-" + uuid.NewString()[:8], nil
}

type MockAI struct{}

func (MockAI) PredictCashflow(_ context.Context, _, _ uuid.UUID, days int) (map[string]float64, error) {
	return map[string]float64{"inflow": float64(days) * 1000, "outflow": float64(days) * 800}, nil
}

func (MockAI) SupplierRisk(_ context.Context, _ uuid.UUID, _ map[string]float64) (float64, error) {
	return 0.25, nil
}
