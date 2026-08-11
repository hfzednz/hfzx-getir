# API Error Codes (NEXORA)

Envelope:

```json
{ "error": { "code": "invalid_argument", "message": "...", "traceId": "...", "retriable": false } }
```

| Code | HTTP | Retriable | Meaning |
|------|------|-----------|---------|
| invalid_argument | 400 | no | Validation failure |
| unauthorized | 401 | no | Auth missing/invalid |
| forbidden | 403 | no | Policy/RBAC denied |
| not_found | 404 | no | Resource missing |
| conflict | 409 | no | Idempotency/state conflict |
| illegal_transition | 409 | no | Workflow state illegal |
| period_closed | 409 | no | Accounting period closed |
| journal_unbalanced | 422 | no | Debits ≠ credits |
| policy_denied | 403 | no | OPA/security deny |
| upstream_failure | 502 | yes | BFF dependency failure |
| internal_error | 500 | yes | Unexpected |
| rate_limited | 409/429 | yes | Rate limit |
| approval_required | 403 | no | LiveOps change needs approval |
| feature_disabled | 403 | no | Flag emergency-off or disabled |
| experiment_closed | 409 | no | Experiment not running |
| supplier_not_verified | 403 | no | Supplier must be approved |
| quality_gate_failed | 422 | no | Required quality gate failed |
| country_inactive | 409 | no | Country not activated for operations |
| rate_limited | 429 | yes | Open platform / gateway rate limit |
| sandbox_violation | 403 | no | Super App plugin permission/signing violation |
| not_ready | 409 | no | Innovation TRL too low to enable |
| gate_failed | 422 | no | Hyperscale or Final Genesis certification gate failed |
