package postgres

import "database/sql"

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
}

func NewRepos(db *sql.DB) *Repos {
	return &Repos{
		Companies: &CompanyRepo{DB: db}, Periods: &PeriodRepo{DB: db}, Accounts: &AccountRepo{DB: db},
		Journals: &JournalRepo{DB: db}, Suppliers: &SupplierRepo{DB: db}, Procurement: &ProcurementRepo{DB: db},
		AP: &APRepo{DB: db}, AR: &ARRepo{DB: db}, Treasury: &TreasuryRepo{DB: db}, Budgets: &BudgetRepo{DB: db},
		Assets: &AssetRepo{DB: db}, Tax: &TaxRepo{DB: db}, Expenses: &ExpenseRepo{DB: db},
		Approvals: &ApprovalRepo{DB: db}, Payroll: &PayrollRepo{DB: db}, Outbox: &OutboxRepo{DB: db},
	}
}
