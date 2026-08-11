# NEXORA Super App Platform

Control plane for modular Super App hosting (WeChat/Grab-style flexibility) while preserving quick-commerce core.

| Item | Value |
|------|--------|
| Service | `superapp-service` `:8121` `/v1/superapp` |
| Flutter shell | `packages/flutter/superapp_shell/` |
| Plugin SDK | `packages/flutter/superapp_plugin_sdk/` |
| Events | `superapp.events` |

## Architecture artifacts

See `services/superapp-service/ARCHITECTURE.md` for:

- Super App / Plugin / Mini App / Extension architecture
- ER diagram
- Folder structure
- Ownership boundaries vs open-platform, liveops, domain SoTs

## Dependency graph

```text
Flutter Shell → superapp-service → LiveOps (enable gate)
                                 → AI (recommendations)
                                 → Kafka outbox
Mini Apps → Extension Hooks → existing BFFs / domain services
Developer apps / API keys → open-platform-service (not duplicated)
```

## Quick start

```bash
cd services/superapp-service && make test && make run
curl -H "X-Tenant-Id: $TENANT" -X POST http://localhost:8121/v1/superapp/bootstrap/mini-apps
curl -H "X-Tenant-Id: $TENANT" "http://localhost:8121/v1/superapp/shell/resolve?subjectId=u1&shellVersion=1.0.0"
```
