# API Optimization Guide

- Enable response compression for payloads >1KB
- Prefer cursor pagination over offset for large lists
- Batch endpoints for cart/notifications where contracts allow
- Retry only on retriable error envelope (`retriable: true`)
- Timeouts: connect 1s, request 5–15s by criticality
- Idempotency keys on mutating payment/order paths (existing)
