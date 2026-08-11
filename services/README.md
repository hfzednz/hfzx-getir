# services/

Go microservices and edge BFFs. Each module has its own `go.mod`, `Makefile`, and (usually) `README.md`.

Bootstrap from repo root via `go.work`.

Registry: `docs/launch/service-registry.yaml`.

DevMode: empty `DATABASE_URL` → in-memory adapters for local tests (not production).
