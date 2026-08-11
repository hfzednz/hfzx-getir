# NEXORA

Enterprise Quick Commerce platform monorepo.

## Source of truth

**[Master Blueprint](docs/constitution/MASTER_BLUEPRINT.md)** — architecture, stack, and roadmap. Constitution wins all conflicts.

**[Monorepo structure](docs/monorepo/STRUCTURE.md)** — canonical on-disk paths (do not rename to prompt examples).

## Clone and run

```bash
git clone <url> nexora && cd nexora

# Windows
pwsh -File scripts/bootstrap.ps1

# macOS / Linux
bash scripts/bootstrap.sh

make doctor
make verify-structure
make test-go-focus   # critical Go services
```

Local data plane:

```bash
docker compose -f infra/docker/docker-compose.yml up -d
```

Web workspaces:

```bash
pnpm install
pnpm --filter admin_web dev
```

Flutter:

```bash
cd apps/mobile_customer && flutter pub get && flutter run
# or: dart pub global activate melos && melos bootstrap
```

Production ops pack: `docs/production/README.md`.

## Layout

| Path | Contents |
|------|----------|
| `apps/` | Mobile + admin web |
| `services/` | Domain + BFF Go services |
| `packages/` | Shared Flutter/web/SDK libraries |
| `infra/` | Terraform, Helm, K8s, Argo, Docker, observability |
| `ops/` | Runbooks, playbooks, release, SLO |
| `docs/` | Constitution, design, API, production, guides |
| `qa/` | Automated quality suites |
| `tools/` | Certifiers / validators |
| `store/` | Store listing / ASO |
| `scripts/` | Bootstrap & verification |

## Workspaces

- Go: `go.work`
- Flutter: `melos.yaml`
- Node: `pnpm-workspace.yaml`

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) → `docs/guides/CONTRIBUTING.md`.

## Status

Prompts **#01–#43** delivered (product, review, production deployment).  
**#44** assembles monorepo glue for clone-build-test-maintain without redesign.
