# Autonomy ops checklist

- [ ] Prometheus scrapes `autonomy-service:8125`
- [ ] Envoy route `/v1/autonomy` → autonomy cluster
- [ ] Kafka topic `autonomy.events` provisioned
- [ ] Genesis certificate issued after city launch soak
- [ ] Heal actions page only when platform-ops execution fails
- [ ] AI CTO reviews archived weekly

Dashboard panels: audit scores, heal success rate, genesis gate panel, evolution backlog age.
