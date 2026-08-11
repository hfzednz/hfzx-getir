# Scaling Guide

1. Apply HPA templates (`infra/hardening/k8s-hpa.yaml`)
2. Pre-warm pools before Black Friday scenario
3. Scale Redis/Kafka partitions before RPS targets
4. Use LiveOps flags for graceful feature shedding
5. Cert gate `scalability` requires all throughput benches pass
