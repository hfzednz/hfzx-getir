# Operations Guide

1. `POST /v1/hyperscale/bootstrap`
2. Review `GET /v1/hyperscale/gates`
3. Run chaos packs in staging (`qa/hyperscale/chaos`)
4. Issue cert: `POST /v1/hyperscale/certificates`
5. Drain outbox; monitor Prometheus job `hyperscale`

Runbooks remain under `ops/runbooks/` — this guide only hardens certification flow.
