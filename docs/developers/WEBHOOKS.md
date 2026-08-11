# Webhooks

## Verification

```
signature = HMAC_SHA256(secret, rawBody)
header X-Nexora-Signature = "sha256=" + hex(signature)
```

## Retries

Exponential backoff up to 5 attempts, then DLQ. Replay via `POST /v1/open/webhooks/deliveries/{id}/replay`.
