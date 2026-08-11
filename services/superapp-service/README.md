# NEXORA Super App Service

Modular Super App control plane: mini-apps, plugins, widgets, store, permissions, shell resolve.

- HTTP `:8121` `/v1/superapp`
- Kafka topic `superapp.events`
- Does **not** own domain SoT, open-platform keys/webhooks, liveops flags, or wallet/loyalty balances

```bash
make test && make run
```
