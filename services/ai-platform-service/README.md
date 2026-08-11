# AI Platform Service

Centralized AI brain for NEXORA. HTTP `:8106` · `/v1/ai` · Python sidecar `:8206`.

```bash
go test ./...
go run ./cmd/ai-platform-service

# optional GPU/ONNX-ready sidecar
cd ml && pip install -r requirements.txt && uvicorn app.main:app --port 8206
```

Does **not** own search, recommendation rails, CRM, notifications, or analytics warehouses.

See [ARCHITECTURE.md](./ARCHITECTURE.md) · [FEATURES.md](./FEATURES.md)
