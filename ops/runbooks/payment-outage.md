# Runbook — Payment outage

## Symptoms

- `PaymentAuthorizeFail` alert
- Checkout failures / order place stuck in payment step

## Mitigate

1. Confirm Stripe/PSP status page + API keys not rotated unexpectedly.
2. LiveOps: enable graceful degrade flag if available (cash-on-delivery / delay messaging) — dual-control.
3. Do **not** enable MockPSP in production.
4. Scale payment-service if CPU/sat; check Redis/DB.
5. Rollback last payment-service image via GitOps if deploy-correlated.

## Escalate

SEV-1 if authorize success < 95% for 10m in a launch city.
