# Production Checklist

## T-7 days

- [ ] RC tagged; CI green (services, infra, quality)
- [ ] Staging soak started
- [ ] Change advisory board notified
- [ ] Store metadata / screenshots frozen
- [ ] Capacity plan reviewed (`ops/production/capacity-planning.md`)

## T-24 hours

- [ ] `go run ./tools/integration-cert` PASS
- [ ] `go run ./tools/prod-validate -env=staging` PASS
- [ ] No Critical CVEs on release images
- [ ] Backup job success < 24h
- [ ] DR DNS drill signed (or waiver within quarter)
- [ ] On-call confirmed; silence windows documented

## T-0 deploy

- [ ] Pre-migration backup / PITR bookmark
- [ ] Migrations expand-only applied
- [ ] GitOps prod PR merged; Argo sync
- [ ] Canary 10→25→50→100 per policy
- [ ] Kill switches dual-control verified
- [ ] Business journey smoke (BUSINESS_VALIDATION.md)

## T+2 hours

- [ ] Error budget within SLO
- [ ] Payment authorize success (excl. declines) within SLO
- [ ] No SEV-1/2 open
- [ ] Mobile crash-free users stable vs baseline

## T+24 hours / T+7 days

- [ ] Post-release review (POST_RELEASE.md)
- [ ] Feature adoption + revenue dashboards reviewed
- [ ] Release certificate signed (CERTIFICATE.md)

Also see: `ops/release/CHECKLIST.md`, `ops/release/GO_NO_GO.md`.
