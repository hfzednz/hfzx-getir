# NEXORA Design System — Index (UI/UX Constitution)

> **Status:** BINDING for all product UI  
> **Version:** 1.0.0  
> **Effective:** 2026-08-06  
> **Authority:** Mandatory under Master Blueprint §§44–47  
> **Rule:** Future prompts MUST NOT invent ad-hoc colors, type, spacing, components, or motion. Use this system.

## Precedence

1. [`docs/constitution/MASTER_BLUEPRINT.md`](../constitution/MASTER_BLUEPRINT.md)
2. **This design system** (`docs/design-system/`)
3. Accepted ADRs
4. Implementation prompts

## Scope

Unified design language for:

| Surface | Density | Primary consumers |
|---------|---------|-------------------|
| Customer App | Comfortable | Flutter `nexora_design` |
| Courier App | Compact | Flutter `nexora_design` |
| Warehouse App | Dense | Flutter `nexora_design` |
| Admin Dashboard | Dense | `@nexora/ui` |
| Super Admin | Dense | `@nexora/ui` |
| Internal Web Panels | Dense / Compact | `@nexora/ui` |

**One token source → many densities.** Never fork a second brand system.

## Document map

| # | Doc | Contents |
|---|-----|----------|
| 01 | [Brand Foundation](01-brand.md) | Philosophy, personality, identity, UX/product principles |
| 02 | [Color System](02-color.md) | Primitives, roles, light/dark, contrast |
| 03 | [Typography](03-typography.md) | Scale, roles, responsive, tabular |
| 04 | [Spacing](04-spacing.md) | 4pt grid, layout rules per surface |
| 05 | [Border & Radius](05-border-radius.md) | Radius, strokes, outlines |
| 06 | [Shadow & Elevation](06-shadow-elevation.md) | Elevation levels, hover |
| 07 | [Iconography](07-iconography.md) | Sizes, styles, catalogs |
| 08 | [Illustration](08-illustration.md) | Empty/success/error states |
| 09 | [Component Library](09-components.md) | Primitives & patterns inventory |
| 10 | [Product Components](10-product-components.md) | Commerce-specific |
| 11 | [Order Components](11-order-components.md) | Cart → tracking → receipt |
| 12 | [Flutter & Web Architecture](12-implementation-architecture.md) | Widget/theme/token impl |
| 13 | [Navigation](13-navigation.md) | Shells, deep links, transitions |
| 14 | [Motion](14-motion.md) | Durations, curves, signatures |
| 15 | [Gestures & Feedback](15-gestures-feedback.md) | Touch, haptics, states |
| 16 | [Loading System](16-loading.md) | Skeleton, shimmer, offline |
| 17 | [Forms](17-forms.md) | Validation, masks, OTP |
| 18 | [Search Experience](18-search.md) | Instant, voice, image, NL |
| 19 | [Map Experience](19-map.md) | Live tracking, zones, ETA |
| 20 | [Accessibility](20-accessibility.md) | WCAG 2.2 AA+, SR, focus |
| 21 | [Responsive System](21-responsive.md) | Breakpoints, adaptive shells |
| 22 | [Figma Structure](22-figma-structure.md) | File architecture |
| — | [Design Tokens JSON](tokens/nexora.tokens.json) | Machine-readable SoT |

## Anti-patterns (forbidden)

- Purple-on-white / purple-indigo gradient themes
- Warm cream + terracotta “AI default” look
- Broadsheet newspaper layouts as product UI
- Generic Material default chrome leaking into apps
- Pill-cluster marketing chrome in first viewport
- Floating badge stickers on hero media
- Cards used only for decoration (constitution: cards only when interaction requires container)
- Placeholder gray boxes as “design”
- Inventing one-off hex values in feature code

## Density profiles

| Token | Customer | Courier | Warehouse | Admin |
|-------|----------|---------|-----------|-------|
| Touch target min | 48px | 48px | 44px (scan-optimized, still ≥44) | 36px desktop / 44 touch |
| List row height | 64–72 | 56–64 | 48–56 | 40–48 |
| Type base | body.md 14 | body.md 14 | body.sm 13 | tableCell 13 |
| Radius preference | md (12) | sm–md | sm (8) | sm–md |
| Motion | Signature allowed | Minimal | Minimal | Micro only |

## Implementation packages

- Flutter: `packages/flutter/nexora_design`
- Web: `packages/web/ui` (`@nexora/ui`) — created in later phase
- Tokens: edit `tokens/nexora.tokens.json` → codegen Dart/CSS/TS

## Compliance checklist (every UI PR)

- [ ] Uses semantic color roles (no raw hex in features)
- [ ] Uses typography tokens
- [ ] Uses spacing/radius/elevation tokens
- [ ] Respects density for the app surface
- [ ] Motion uses `NxDuration` / `NxCurves` + reduced-motion path
- [ ] Contrast AA verified for text/icon on surfaces
- [ ] Focus/semantics present for interactive controls
- [ ] No new primitive without DS doc update
