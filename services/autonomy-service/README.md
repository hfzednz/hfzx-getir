# NEXORA Autonomy Service

Autonomous enterprise delivery, self-evolution, and Final Genesis certification.

- HTTP `:8125` `/v1/autonomy`
- gRPC stub `:9125`
- Kafka topic `autonomy.events`
- Does **not** redesign services or replace platform-ops/quality/security/hyperscale SoT

```bash
make test && make run
```
