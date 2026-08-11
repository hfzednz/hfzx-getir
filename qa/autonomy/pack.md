# Autonomy QA pack

| Suite | Intent |
|-------|--------|
| unit | `go test ./...` in autonomy-service |
| genesis | Bootstrap → gates → issue certificate |
| heal | Heal plan records + platform-ops mock |
| regression | Gate failure when hyperscale port unhealthy |

CI: run `make test` in `services/autonomy-service`.
