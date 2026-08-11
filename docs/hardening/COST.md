# Cost Optimization Guide

- Right-size requests/limits via HPA (65% CPU target)
- Spot/preemptible for non-critical batch/analytics workers
- Compress Kafka (zstd) + shorter retention on high-volume topics
- GPU only for hot AI models; CPU for light ranking
- Storage lifecycle: hot→warm→cold for logs/events
