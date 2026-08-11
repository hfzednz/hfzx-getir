# NEXORA Developer Onboarding

## Day 0 — Context

1. Product vision & constitution: `docs/constitution/MASTER_BLUEPRINT.md`
2. Design system index: `docs/design-system/00-INDEX.md`
3. Brand: `docs/design-system/01-brand.md`
4. Diagrams: `docs/architecture/diagrams/`

## Day 1 — Environment

1. Install: Go (≥ version in `go.work`), Flutter, Node (LTS), pnpm, Docker, protobuf/buf tooling
2. Bootstrap: `pwsh -File scripts/bootstrap.ps1` or `bash scripts/bootstrap.sh`
3. Read monorepo map: `docs/monorepo/STRUCTURE.md`
4. Start local deps: `docker compose -f infra/docker/docker-compose.yml up -d`
5. Run customer app steps in `apps/mobile_customer/README.md`
6. Confirm access to package registries & cloud (as provided by platform team)

## Day 2 — Make a safe change

1. Pick a small docs or lint task
2. Open PR following `docs/guides/CONTRIBUTING.md`
3. Watch CI path filters for your area

## Mental model

- **City** is the scale/tenant dimension
- **BFF** shapes client APIs
- **Kafka** carries facts between domains
- **Postgres** is system of record
- **Offline outbox** is mandatory for field apps

## Who to ask

| Area | Owner path |
|------|------------|
| Mobile | `apps/mobile_*`, `packages/flutter/*` |
| Backend | `services/*` |
| Admin | `apps/admin_web` |
| Infra | `infra/*` |
| AI | `ai/*` |

Use CODEOWNERS for review routing.
