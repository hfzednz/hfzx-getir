# Business Validation Matrix

Execute after canary ≥ 10% and again at 100%.

| Journey | Actor | Steps | Pass |
|---------|-------|-------|------|
| Customer order | Customer app / BFF | OTP → home → cart → preview → place → track | orderId + tracking status |
| Payment | System | authorize (real PSP staging/prod) | IntentAuthorized |
| Courier | Courier app | offer → accept → pickup → deliver | status transitions |
| Warehouse | Warehouse app | queue → pick → pack → handoff | task complete |
| Admin | admin_web | list order → detail → audited action | dual-control where required |
| Refund | Admin + payment | refund minor units | Refund succeeded |
| Notification | System | push/SMS/email on place | delivery accepted by provider |
| Kill switch | Admin + LiveOps | toggle flag dual-control | LiveOps SoT, no fail-open |

Automated assist: `tools/prod-validate` + quality-service release cert.
