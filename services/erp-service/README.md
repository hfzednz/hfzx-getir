# NEXORA ERP Service

Enterprise finance, accounting, procurement, treasury, budgeting, tax packs, fixed assets, and approval workflows.

- HTTP `:8108` base `/v1/erp`
- gRPC stub `:9108`
- Memory mode when `DATABASE_URL` is empty
- Posts journals via `finance-ledger-service` port (does not own payment/wallet/settlement execution or inventory stock SoT)

## Run

```bash
make test
make run
```

## Key flows

1. Company → fiscal year/periods → COA → balanced journals (ledger port)
2. Supplier → PO → GRN (inventory receive port) → AP 3-way match → approve → schedule payment (settlement port)
3. Budgets / expenses / multi-level approvals
4. Assets + straight-line depreciation journals
5. Tax return calculation packs

See `ARCHITECTURE.md` and `FEATURES.md`.
