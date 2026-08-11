# Security Service Features

## Zero Trust & Policy
- Adaptive trust scoring (identity + device + context)
- Device attestation / root / jailbreak / tamper signals
- OPA-backed policy evaluate (access/data/compliance/security/runtime/AI)
- JIT temporary access requests with TTL approval (not IAM role SoT)

## Secrets & Audit
- Secret metadata + Vault rotate / cert renew (no plaintext storage)
- Immutable hash-chained audit ledger with search

## Threat / Vuln / IR
- Threat ingest (ATO, bot, stuffing, brute force, priv-esc heuristics)
- DevSecOps findings (SAST/DAST/deps/container/IaC/secret)
- Incidents with SOAR playbook stub, timeline, postmortem close

## Compliance & Data Governance
- Framework controls (GDPR/KVKK/PCI/ISO/SOC2/OWASP/NIST/CIS)
- Evidence attach + compliance audit score
- Data classification / PII tags / privacy export|erase|consent

## Risk / Fraud / AI Security
- Risk register scoring
- Fraud signals via fraud-service facade port
- Prompt injection detection + AI guardrail port

## Platform
- Outbox on `security.events`, admin posture stats, rate limits, OpenAPI/proto/i18n
