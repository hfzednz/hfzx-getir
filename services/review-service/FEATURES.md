# Review Service — Feature Matrix

| Area | Capability | Status |
|------|------------|--------|
| Reviews | Text / anonymous / verified / edit history / soft-delete | ✅ |
| Ratings | stars_5 / emoji / thumbs + Bayesian + time-decay | ✅ |
| Media | Refs via media port (image/video/voice) | ✅ |
| Interactions | Helpful votes, comments, reports, pin | ✅ |
| Moderation | Heuristic + AI port + human queue/decide | ✅ |
| Fraud | Dup body, velocity, review-bomb heuristics | ✅ |
| Trust | Verified buyer, badges, AI trust weight | ✅ |
| Reputation | Per-target score + tier | ✅ |
| Quality | Multi-dimension scores on review | ✅ |
| AI | Sentiment, topics, summarize, unsafe labels | ✅ |
| Search | In-memory / OpenSearch adapter | ✅ |
| Events | Outbox + Kafka stub | ✅ |
| Admin | Moderation queue, stats | ✅ |
| Compliance | GDPR/KVKK delete (content wipe) | ✅ |
| Security | Rate limit, tenant isolation, PII mask hook | ✅ |

## Admin surfaces (API-backed)

Moderation Dashboard · Review Explorer · Trust Dashboard · Spam signals · Sentiment/Quality/Reputation reads via aggregates + AI summarize.
