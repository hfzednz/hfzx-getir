# Security Report (Prompt-42)

## Remediated

| Issue | Fix |
|-------|-----|
| Fake PSP authorize on Stripe failure | Mock removed from production failover chain |
| Default OTP pepper in production | Required env; refuse `nexora-otp-pepper` when DB set |
| OTP_DEV_MODE logging codes in prod | Forbidden when `DATABASE_URL` set |
| Ephemeral JWT with durable sessions | `JWT_KEY_PEM` required with DB |
| Over-broad CORS `*` with DB | Config hard-fail |
| Kill-switch inventing success | Fail closed to client |

## Residual risk

- DevMode still uses memory repos + optional mock PSP (local only)  
- Social IdP userinfo mapping may be generic for Facebook/GitHub  
- Certificate pinning / full OWASP sweep of all HTTP handlers not completed this wave  
- Dependency CVE scan not run this wave (`govulncheck` / npm audit pending)  
