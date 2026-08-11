# NEXORA Test Strategy

## Principles

- Shift-left: unit + contract on every PR
- Shift-right: synthetic probes + SLO burn
- Risk-based DoD: higher blast radius → more layers
- No business logic in QA platform — only validation

## Pyramid

| Layer | Owner path | Gate |
|-------|------------|------|
| Unit | `services/*/...` | coverage ≥ 70% lines (critical services higher) |
| Integration | `tools/integration-cert`, service IT | 0 failed |
| API | `qa/suites/api` | contract + smoke |
| UI | `qa/playwright`, Flutter `integration_test` | critical journeys green |
| E2E | `qa/suites/e2e/*` | customer/courier/warehouse/admin |
| Perf | `qa/k6` | p95 / error-rate thresholds |
| Security | `qa/zap` + SAST/deps | 0 critical |
| A11y | `qa/suites/a11y` | WCAG AA blockers = 0 |
| Chaos | `qa/chaos` | recovery SLOs met |
| AI | `qa/suites/ai` | latency + safety checks |

## Environments

local → CI PR → nightly → staging soak → prod synthetic

## Certification workflow

1. Bootstrap suites/policies (`POST /v1/quality/bootstrap`)
2. Execute suites; ingest results/coverage/perf/security
3. Evaluate gates
4. Issue certification (`release_readiness` → `production`)
5. Block deploy if `QualityGateFailed`
