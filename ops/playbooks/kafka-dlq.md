# Playbook — Kafka DLQ drain

1. List DLQ depth per topic.
2. Sample payloads; classify poison vs transient.
3. Transient: replay with backoff; bump consumer.
4. Poison: quarantine + ticket to owning service; do not infinite-retry.
5. Record counts in incident timeline.
