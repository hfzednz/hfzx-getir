# settlement-service — Features

| Feature | Status | Notes |
|---------|--------|-------|
| CreateBatch | ✅ | Draft; idempotent |
| AddLine | ✅ | payee_type: courier\|supplier\|merchant\|partner |
| Submit | ✅ | draft → pending_approval |
| Approve | ✅ | Dual-control stub (different actor) |
| ExecutePayouts | ✅ | Ledger + Payout ports; → completed |
| ReconcileProviderReport | ✅ | Writes mismatch when totals differ |
| List / GetBatch | ✅ | Tenant-scoped |
| Outbox + PublishPending | ✅ | `settlement.lifecycle` |
| OpenAPI + Proto | ✅ | Contracts under `api/` / `proto/` |
| Postgres migrations | ✅ | batches, lines, payouts, reconcile, outbox |
| Postgres runtime adapters | 🔶 | Migrations ready; memory repos in dev |
| Real Ledger/Payout HTTP | 🔶 | Stubs succeed; wire HTTP later |

## Non-goals

- Cart / order / inventory aggregates
- Float money
- Same-actor approval of settlements
