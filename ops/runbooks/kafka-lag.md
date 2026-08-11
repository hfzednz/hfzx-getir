# Runbook — Kafka consumer lag

## Symptoms

- `KafkaConsumerLag` alert
- Delayed notifications / search / outbox projection

## Mitigate

1. Identify consumer group + topic.
2. Scale consumers (KEDA / HPA) if lag climbing.
3. Check poison messages → DLQ (`ops/playbooks/kafka-dlq.md`).
4. Pause non-critical consumers to protect order/payment path.
5. After recover, verify lag trend down 15m.
