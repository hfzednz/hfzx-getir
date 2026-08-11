# Auto Scaling

## Horizontal (HPA)

Helm defaults (`infra/helm/nexora/values.yaml`):

- `hpa.enabled: true`, CPU 70%, min 2, max 20 (service overrides for BFFs/orders).

Prod values (`values/prod.yaml`): raise max for order, payment, bff-customer, realtime-gateway.

## Cluster Autoscaler

Terraform prod: `node_min=6`, `node_max=80`, GPU pool enabled.

## KEDA

- Order path: `infra/k8s/base/keda-order.yaml` (Kafka lag / queue depth triggers).
- Extend similarly for notification workers and AI inference queues without changing service code.

## Predictive / scheduled

- LiveOps calendar + platform-ops capacity hooks for known peaks (lunch/dinner).
- Load env used to calibrate before city launch.

## GPU

- AI platform / innovation inference on GPU node pool; scale on queue depth + GPU util alerts.
