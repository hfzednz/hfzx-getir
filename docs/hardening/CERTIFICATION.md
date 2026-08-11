# Hyperscale Production Certification

## Certificate

Issued by `hyperscale-cert-service` when **all** gates pass:

- performance · security · scalability · reliability  
- observability · disaster_recovery · zero_critical_findings  

Validity: **90 days**. Regression (e.g. API p99 >150ms) blocks re-issue.

## Quality statement

| Gate | Requirement |
|------|-------------|
| Zero critical bugs | Open critical findings = 0 |
| Zero critical vulns | security-service port green |
| Zero release blockers | quality-service gates green |
| Availability | 99.99% target tracked in SLO catalog |
| DR | Chaos + platform-ops DR drill |

## Artifact locations

- Audits: `docs/hardening/00-AUDIT-AND-GAPS.md`
- Configs: `infra/hardening/`
- Benches: `qa/hyperscale/`
- Service: `:8124` `/v1/hyperscale`
