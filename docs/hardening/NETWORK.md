# Network Optimization Guide

See `infra/hardening/envoy-http3.yaml`:

- HTTP/3 at CDN edge; HTTP/2 to mesh
- Brotli/gzip for JSON
- LEAST_REQUEST LB + outlier ejection
- TLS1.2+ ; short DNS TTL for failover
