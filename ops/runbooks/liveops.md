# LiveOps runbook

## Flag emergency disable

1. `POST /v1/liveops/rollbacks` with `{ "kind":"emergency", "subjectKey":"<flagKey>", "reason":"..." }` or upsert flag with `emergencyOff:true`.
2. Confirm `FeatureDisabled` / `RollbackExecuted` on `liveops.events`.
3. Invalidate edge caches if BFFs cache evaluations.

## Experiment auto-rollback

1. If primary metric / crash rate guardrail trips, complete experiment without winner or rollback flag tied to treatment.
2. `POST /v1/liveops/rollbacks` `{ "kind":"experiment", "subjectKey":"<expKey>" }`.

## Config rollback

`POST /v1/liveops/rollbacks` `{ "kind":"config", "subjectKey":"<configKey>", "toVersion":N }`.
