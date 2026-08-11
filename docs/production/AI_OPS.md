# AI Operations

## Model deployment

1. Register artifact in `ai-platform-service` model registry.
2. Shadow → canary → primary routing (existing inference ports).
3. Monitor latency, error, drift score.
4. Rollback = route traffic to previous model id (no domain redesign).

## Prompt deployment

- Versioned prompt packs via AI platform; dual-control for customer-facing prompts.
- Guardrails enforced by security-service prompt policies.

## Monitoring

| Signal | Alert |
|--------|-------|
| Inference p99 | AIInferenceLatency |
| Drift | threshold breach → ticket + optional auto-fallback |
| GPU util | scale / queue backlog |
| Fallback rate | high → page if > 20% 15m |

## GPU

- Prod GPU node pool (Terraform). Scale with KEDA on inference queue depth.
