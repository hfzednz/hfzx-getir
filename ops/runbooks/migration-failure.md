# Runbook — Migration failure

## Symptoms

- Deploy blocked; migrator non-zero exit
- Schema version mismatch vs expected

## Mitigate

1. Halt Argo sync / canary for dependent services.
2. Inspect migrator logs; determine expand vs destructive.
3. If expand partially applied: prefer forward-fix compensating migration.
4. If destructive/corrupt: PITR to `pre-release-<tag>` bookmark (`docs/production/DATABASE_RELEASE.md`).
5. Re-run `tools/prod-validate` before resume.

## Escalate

CISO + CTO for any PITR on production ledger data.
