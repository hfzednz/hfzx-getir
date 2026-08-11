# Go / No-Go Checklist — Global Launch

## Mandatory (all must be YES)

- [x] Domain services #09–#29 present with ARCHITECTURE.md
- [x] Edge BFFs present (`bff-customer` `:8111`, `bff-courier` `:8112`, `bff-warehouse` `:8113`, `bff-admin` `:8114`)
- [x] `realtime-gateway` `:8115` SSE/publish fanout
- [x] Infra GitOps (`infra/`, Argo CD, Helm, Kustomize)
- [x] Security GRC + platform-ops control plane
- [x] Service registry committed (`docs/launch/service-registry.yaml`)
- [ ] `go run ./tools/integration-cert` PASS on release candidate SHA
- [ ] `go run ./tools/prod-validate -env=staging` PASS
- [ ] Staging soak ≥ 24h with error budget within SLO
- [ ] DR DNS failover drill signed
- [ ] No open Critical CVEs on release images
- [ ] Dual-control kill switches verified via bff-admin
- [ ] Production secrets: `JWT_KEY_PEM`, `OTP_PEPPER`, PSP keys; `OTP_DEV_MODE=false`
- [ ] Mobile tracks ready (Play internal+/TestFlight) if client release included
- [ ] Canary policy acknowledged (`infra/k8s/overlays/prod/canary-analysis.yaml`)
- [ ] On-call primary confirmed (`ops/production/oncall.md`)

## Go criteria

**GO** when integration-cert PASS + staging soak green + DR drill signed.  
**NO-GO** on payment/checkout/identity regression or Critical CVE.

## Sign-off

| Role | Name | Date | Decision |
|------|------|------|----------|
| CTO | | | |
| CISO | | | |
| SRE Lead | | | |
| Release Manager | | | |
