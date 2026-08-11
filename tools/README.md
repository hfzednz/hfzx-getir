# tools/

| Tool | Purpose |
|------|---------|
| `integration-cert` | Launch integration certification |
| `prod-validate` | Environment health / structural release checks |
| `genesis-cert` | Autonomy genesis gates |
| `hyperscale-cert` | Hyperscale certification helper |

```bash
go run ./tools/prod-validate -env=staging
go run ./tools/integration-cert
```
