# NEXORA Quality Service

Quality gates, test run registry, release certification.

- HTTP `:8118` `/v1/quality`
- Suites live under repo `qa/` — this service does **not** own product business logic

```bash
make test && make run
```
