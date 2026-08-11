# ERP Service Features

## Accounting
- Multi-company books, fiscal years, open/close periods
- Chart of accounts (asset/liability/equity/revenue/expense)
- Balanced journal entries (int64 minor units) with idempotency
- Period-closed guard; posts to finance-ledger port

## Procurement & AP
- Suppliers with AI risk score hint
- Purchase requests → approval start
- Purchase orders, goods receipts (inventory receive port)
- AP invoices with 3-way match score; approve; schedule supplier payment (settlement port)

## AR / Treasury / Budget
- Corporate AR invoices
- Bank accounts, txn import, reconcile flag, cashflow AI forecast
- Annual/quarterly/monthly budgets + approve + variance lines

## Assets / Tax / Expenses / Payroll
- Fixed asset registry + straight-line depreciation journals
- VAT/corporate/withholding tax return packs
- Expense reports + approval workflow
- Payroll batch export stub (external payroll ref)

## Platform
- Outbox events on `erp.events`
- Admin stats, rate limits, tenant middleware, OpenAPI + proto stubs
- IFRS-oriented COA/period model; money never float
