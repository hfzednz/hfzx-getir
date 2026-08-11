# Security Hardening Guide

- Attack surface: only Envoy-exposed routes; no direct pod ingress
- OWASP ASVS baseline + ZAP in quality suites
- Dependency pins + weekly advisory scan (`ops/hardening/dependency-audit.md`)
- Secret/key rotation via Vault metadata in security-service
- Verify Zero Trust adaptive signals remain green before cert issue
- No redesign of IAM/PSP ownership
