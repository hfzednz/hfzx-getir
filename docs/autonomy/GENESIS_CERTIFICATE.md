# Final Genesis Certificate

**Issuer:** `autonomy-service` `:8125`  
**Topic:** `autonomy.events` → `GenesisCertified`

## Required gates

| Gate | Meaning |
|------|---------|
| `autonomy_audits` | ≥11 completed scope audits |
| `self_healing` | ≥3 executed heal actions |
| `ai_cto` | ≥3 AI CTO reviews |
| `evolution` | ≥1 evolution task |
| `release_engine` | release score ≥80 |
| `governance` | ≥6 healthy governance loops |
| `hyperscale` | hyperscale-cert port certified |
| `security` | security port healthy |
| `quality` | quality port healthy |

## Issue

```bash
curl -X POST -H "X-Tenant-Id: $TENANT" -d '{"version":"1.0.0"}' \
  http://localhost:8125/v1/autonomy/genesis
```

When all gates pass, the ecosystem is **enterprise-certified** for country-specific configuration-only deployment. No architectural additions required.
