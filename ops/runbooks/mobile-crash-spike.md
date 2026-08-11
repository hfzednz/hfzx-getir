# Runbook — Mobile crash spike

## Symptoms

- `MobileCrashSpike` / Crashlytics spike
- Store rating drops

## Mitigate

1. Identify app + version (+BUILD).
2. Halt Play staged rollout / pause App Store phased release.
3. LiveOps min-version or kill feature flag if crash tied to flag.
4. Hotfix build → internal → accelerated production.
5. Post note in `ops/release/notes/`.
