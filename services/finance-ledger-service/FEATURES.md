# finance-ledger-service — Features

| Feature | Status | Notes |
|---------|--------|-------|
| EnsureAccount | ✅ | Idempotent by tenant+code |
| PostJournal | ✅ | Rejects unbalanced; immutable after post |
| GetBalance | ✅ | debit−credit from posted lines |
| ListJournals / GetJournal | ✅ | Tenant-scoped |
| CreateInvoice | ✅ | Tax via TaxRule / default tax code |
| IssueCreditNote | ✅ | Full credit → invoice `credited` |
| TaxCalculate | ✅ | Integer bps math, round half-up |
| UpsertTaxRule | ✅ | Basis points 0..10000 |
| Outbox + PublishPending | ✅ | `ledger.lifecycle` |
| OpenAPI + Proto | ✅ | Contracts under `api/` / `proto/` |
| Postgres migrations | ✅ | COA, journals, invoices, tax, outbox |
| Postgres runtime adapters | 🔶 | Migrations ready; memory repos in dev |
| Kafka real producer | 🔶 | Stub publisher |

## Non-goals

- Cart / order / inventory aggregates
- Loyalty points engine
- Float money
- Cross-service FK constraints (opaque refs only)
