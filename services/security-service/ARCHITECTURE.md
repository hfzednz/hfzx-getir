# NEXORA Security Service — Enterprise Security, Compliance, Governance & Risk

> Binding under Master Blueprint §7 (`security-service`).  
> Stack: Go · PostgreSQL · Redis · Kafka · Vault port · OPA policy · OTel · Prometheus.  
> **Hard rules:** Does **not** own IAM authn/sessions/credentials (`identity-service`), payment PSP/PAN/3DS (`payment-service`), or wallet crypto.  
> Coordinates audit, threat, secrets metadata, compliance evidence, risk, incidents. `audit-service` / `fraud-service` remain split-ready facades via ports.

## Mission

Zero Trust continuous verification signals, policy decisions (OPA), immutable audit, secrets lifecycle (Vault), vulnerability & DevSecOps findings, incident response, compliance packs (GDPR/KVKK/PCI/ISO/SOC2), data governance, risk register, AI prompt/guardrail security.

## Architecture

```mermaid
flowchart LR
  Admin --> SEC[security-service :8109]
  SEC --> Vault[Vault port]
  SEC --> OPA[OPA evaluate]
  SEC --> IAM[identity trust signals]
  SEC --> AI[ai-platform guardrails]
  SEC --> SIEM[SIEM/SOAR ports]
  SEC --> Outbox --> Kafka
```

## Boundaries

| Owns | Does not own |
|------|----------------|
| Security policies, audit ledger, incidents | Login/password/OTP issuance |
| Secrets metadata + rotation orchestration | Raw secret values at rest (Vault) |
| Vulnerability/scan findings | Payment fraud capture SoT |
| Compliance evidence & data classification | Catalog / OMS / Inventory |
| Risk register & vendor risk | Feature flags product SoT |
| Threat alerts & playbooks | Network WAF appliance config |

## Folder structure

```text
services/security-service/
  ARCHITECTURE.md README.md FEATURES.md
  cmd/security-service/
  internal/{config,domain,app,adapters,...}
  migrations/ api/ policies/
```

## API (`:8109` `/v1/security/...`)

Policies · evaluate · audits · secrets · threats · vulns · incidents · compliance · data-gov · risks · access requests · AI security · DevSecOps · admin · outbox

## Events

`SecurityAlertCreated` · `ThreatDetected` · `PolicyViolated` · `SecretRotated` · `CertificateRenewed` · `IncidentOpened` · `IncidentClosed` · `ComplianceAuditCompleted`

## Dependency graph

```mermaid
flowchart LR
  SEC --> Vault
  SEC --> OPA
  SEC --> Identity
  SEC --> AI
  SEC --> FraudFacade
  Payment -.->|no ownership| SEC
  IAM -.->|signals only| SEC
```

## ER (logical)

```mermaid
erDiagram
  POLICY ||--o{ POLICY_EVAL : evaluates
  AUDIT_EVENT ||--|| TENANT : scoped
  SECRET_META ||--o{ ROTATION : rotates
  THREAT_ALERT ||--o{ INCIDENT : escalates
  INCIDENT ||--o{ PLAYBOOK_RUN : responds
  COMPLIANCE_CONTROL ||--o{ EVIDENCE : proves
  RISK_ITEM ||--o{ RISK_SCORE : scored
  SCAN_FINDING ||--o{ REMEDIATION : tracks
```
