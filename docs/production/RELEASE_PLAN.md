# Release Plan

## Versioning

- **Services:** Semantic Versioning `MAJOR.MINOR.PATCH` + Git SHA suffix in image labels.
- **Mobile:** `pubspec.yaml` `version: X.Y.Z+BUILD` — `X.Y.Z` = store versionName; `BUILD` = versionCode / CFBundleVersion (monotonic).
- **Git tags:** `vYYYY.MM.DD` production; `vYYYY.MM.DD-rcN` candidates; `mobile/customer/X.Y.Z+N`.

## Promotion path

```text
main → CI (ci-services, ci-infra, ci-quality)
    → images → GHCR
    → GitOps PR staging (cd-gitops)
    → Argo sync staging → soak ≥ 24h
    → integration-cert + prod-validate
    → Go/No-Go (ops/release/GO_NO_GO.md)
    → GitOps PR prod + canary
    → full weight → post-release watch 2h / 24h / 7d
```

## Freeze windows

- No prod deploys during peak dinner (local TZ) unless SEV-1.
- Schema migrations: maintenance window or expand/contract only (see DATABASE_RELEASE.md).

## Changelog

- Conventional Commits → `.github/workflows/release-changelog.yml` generates notes into `ops/release/notes/`.
- Mobile store “What’s New” sourced from `store/aso/*/whats-new.md`.

## Ownership

| Gate | Owner |
|------|-------|
| Image build | CI |
| Staging promote | Release engineer |
| Prod GO | CTO + CISO + SRE + Release Manager |
| Store submit | Mobile release owner |
| Flag rollout | LiveOps dual-control |
