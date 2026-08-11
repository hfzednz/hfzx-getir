# ADR-0001: Adopt NEXORA Design System as UI Source of Truth

- **Status:** Accepted
- **Date:** 2026-08-06
- **Constitution refs:** §§44–47, Master Prompt 02
- **Deciders:** Design System Architect / Frontend Technical Lead

## Context

Multiple client surfaces (customer, courier, warehouse, admin, super admin) require a unified visual and interaction language without forking Material defaults or inventing per-app palettes.

## Decision

Adopt `docs/design-system/` (index `00-INDEX.md`) and `tokens/nexora.tokens.json` as the mandatory UI/UX source of truth. Implement via `nexora_design` (Flutter) and `@nexora/ui` (web). Brand: Kinetic Clarity — teal `#0B6E6E`, citrus `#E8F07A`, graphite neutrals, Satoshi + Geist.

## Alternatives considered

1. Per-app Material theming — rejected (inconsistent, generic)
2. Multiple brand systems by surface — rejected (ecosystem fragmentation)

## Consequences

### Positive
- Consistent premium QC experience; shared tokens; density profiles scale ops vs retail

### Negative
- Requires discipline and codegen; designers/eng must update DS before one-off UI

### Follow-ups
- Token codegen pipeline in CI
- Widgetbook / Storybook gates
- Contrast CI on role pairs

## Compliance

Master Blueprint §§44–45 updated to reference DS index.
