# NEXORA Coding Guidelines

> Binding under Master Blueprint §§12–16. Amendments require ADR.

## Principles

1. Production-grade only — no placeholder business logic in main paths.
2. Explicit over clever.
3. Boundaries respected (clients → BFF → services).
4. Money never as floating point.
5. Every side effect is idempotent or safely retryable.

## Go

- Package layout: `cmd/`, `internal/`, `api/` (proto stubs), `migrations/`
- Return `error` with wrapped context (`fmt.Errorf("…: %w", err)`)
- Validate at boundary; domain assumes valid commands
- Use `context.Context` cancellation
- Metrics/traces on every external call

## Flutter

- Feature-first clean architecture
- Presentation → domain ← data
- Riverpod for DI/state
- Dio only inside data sources
- l10n for all user strings
- Design tokens via `nexora_design` only

## TypeScript / React

- `strict: true`
- Prefer server components / clear data boundaries in admin
- No direct domain DB access from web

## Reviews

- Security-sensitive: identity, payment, wallet, flags — 2 reviewers
- Performance-sensitive: dispatch, search, tracking — include perf checklist
