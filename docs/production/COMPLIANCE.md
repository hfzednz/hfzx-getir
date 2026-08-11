# Compliance

## GDPR / KVKK

- Lawful basis + consent via identity/profile preferences.
- Data residency hints from `global-service`; enforce storage region in Terraform data module.
- Erasure / export: data-platform redaction events + CRM/support runbooks.
- DPIA for new processors before GA.

## PCI-DSS

- SAQ / ROC as contracted; NEXORA never persists raw PAN/CVV.
- Payment-service is in PCI scope boundary; admin tools must not log tokens.

## Store policies

- Privacy nutrition labels / Data safety forms: `docs/production/mobile/` + `store/aso/*/privacy.md`.
- Age rating, permissions justification (location, notifications, camera) documented.

## Accessibility

- Mobile/web WCAG 2.1 AA target; quality-service a11y suites in QA.

## Retention

| Class | Default retention |
|-------|-------------------|
| Audit logs | ≥ 1 year (security policy) |
| Metrics | 45d prod (Terraform obs module) |
| Backups | 35d prod |
| Chat/support | per CRM policy |
| Order financial | per finance/legal pack |
