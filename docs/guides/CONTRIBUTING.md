# Contributing to NEXORA

## Before you start

1. Read `docs/constitution/MASTER_BLUEPRINT.md` — it wins all conflicts.
2. Read design system: `docs/design-system/00-INDEX.md` and brand `01-brand.md`.
3. Skim ADRs in `ADR/`.
4. UI work must use design tokens/components — no ad-hoc hex/type/spacing.

## Workflow

1. Create branch: `feat/<area>-<summary>`, `fix/…`, `chore/…`
2. Keep PRs scoped to one prompt/phase slice when possible
3. Update contracts (OpenAPI/proto) in the same PR as code
4. Include tests proportional to risk
5. Update docs when behavior or architecture changes
6. Add ADR for cross-cutting decisions

## Commit messages

Prefer conventional commits:

- `feat(order): add cancel window guard`
- `fix(inventory): release reservation on payment failure`
- `docs(constitution): clarify city sharding`

## Definition of Done

- [ ] Constitution-compliant
- [ ] Lint/tests green
- [ ] Observability fields present on new endpoints
- [ ] Migrations reversible or expand/contract safe
- [ ] No secrets committed
- [ ] Feature flags for risky rollouts

## Code of conduct (engineering)

Be precise, kind, and operationally honest — same as the product voice.
