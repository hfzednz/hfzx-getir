# Search Service

Enterprise product discovery. HTTP `:8104` · `/v1/search`.

```bash
go test ./...
go run ./cmd/search-service
```

Does **not** own catalog, promotions, or recommendation SoT (calls recommendation port).

See [ARCHITECTURE.md](./ARCHITECTURE.md) · [FEATURES.md](./FEATURES.md)
