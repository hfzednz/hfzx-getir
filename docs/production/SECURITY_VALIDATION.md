# Security Validation (Pre-Prod)

## Mandatory gates

- [ ] TLS 1.2+ everywhere; HSTS on edge
- [ ] Certificates valid > 14 days (cert-manager alerts)
- [ ] Secrets only via Vault/ExternalSecrets — no plaintext in Git
- [ ] WAF rules enabled (OWASP CRS baseline)
- [ ] NetworkPolicies deny-by-default between namespaces
- [ ] Rate limits on identity OTP and public APIs
- [ ] Zero Trust policies loaded (`security-service`)
- [ ] No MockPSP / OTP_DEV_MODE / wildcard CORS in prod config
- [ ] `govulncheck` + image scan: zero Critical on release images
- [ ] PCI scope: PAN never stored; PSP tokenization only

## Penetration / hardening

- Evidence from Prompt-39 hyperscale/hardening packs.
- ZAP baseline in `ci-quality.yml` must be non-blocking only for informational; Critical blockers fail release.

## Break-glass access

- JIT via security-service approvals; session time-boxed; audit hash-chain.
