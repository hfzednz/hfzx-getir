# Quality engineering runbook

## Gate failure

1. Inspect `GET /v1/quality/admin/stats` and failed run.
2. Fix failing suite; re-ingest results.
3. Re-evaluate gates; re-issue certification.

## Flaky quarantine

Track via `GET /v1/quality/flaky`; quarantine in CI retries=1 then skip with ticket.
