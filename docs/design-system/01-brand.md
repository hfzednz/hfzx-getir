# 01 — Brand Foundation

> Part of the binding Design System — see [00-INDEX.md](00-INDEX.md).

## Design Philosophy

**Kinetic Clarity** — NEXORA interfaces prioritize decision speed in urban minutes-matter contexts. Visual systems are engineered so users always know: *where am I, what can I do next, how long will it take, what will it cost.*

Principles behind the philosophy:

| Principle | Meaning |
|-----------|---------|
| Precision over personality theater | Brand is strong but never loud for its own sake |
| Time as a first-class UI object | ETA, SLA remainder, and courier proximity are hierarchy equals to price |
| Honesty under pressure | Stock, fees, and delays are surfaced early |
| Instrument, not magazine | Dashboards and ops apps read like tools; customer app reads like a confident retailer |
| One job per section | Constitution-aligned composition |

## Brand Personality

| Trait | Expression | Anti-pattern |
|-------|------------|--------------|
| **Precise** | Alignments, tabular numbers, crisp icons | Sloppy spacing, rounded-pill chaos |
| **Urban** | Cool graphite, teal ink, city photography | Pastoral clichés, fake rustic |
| **Kinetic** | Purposeful motion, live tracking | Gratuitous parallax |
| **Trustworthy** | Clear fees, calm errors | Dark patterns, hidden costs |
| **Human** | Warm microcopy, respectful empty states | Robotic error codes as UX |

**Voice:** Direct, calm, slightly sharp. Prefer “Arrives in 12 min” over “Your delightful order is zooming!”

## Visual Identity

### Wordmark

- Logotype: **NEXORA** in Satoshi Bold / ExtraBold, tracked −1% to −2%
- Lockup clear space: ≥ 0.5× cap-height on all sides
- Minimum digital size: 16px cap-height (mobile), 20px (desktop header)
- Do not outline, add gradients to glyphs, or place on busy photo without scrim

### Mark (optional app icon)

- Geometric “N” cut from a rounded square (radius = `radius.md` scaled)
- Field: `color.brand.primary` (`#0B6E6E`)
- Accent spark: single citrus bar `#E8F07A` (not a sticker cluster)

### Photography

- Real product, real dark-store ops, real city night/day
- Full-bleed on marketing/hero surfaces
- Avoid stock “happy fridge” clichés when authentic ops imagery exists
- Overlays: max 40% ink scrim for text legibility; no floating badge stickers

### Graphic language

- Soft geometric grids, 12px radius language
- Signal citrus used sparingly for primary CTAs and live indicators
- Teal for brand chrome and selected states
- No purple gradient identity system

## Emotional Goals

| Moment | Emotion to evoke |
|--------|------------------|
| First open / home | Confidence (“this city is covered”) |
| Search & browse | Momentum without anxiety |
| Checkout | Control and clarity |
| Tracking | Reassurance + anticipation |
| Issue/refund | Respect and competence |
| Courier peak hours | Focus, not panic |
| Warehouse scan | Flow state accuracy |
| Admin incident | Command presence |

## UX Principles

1. **Time honesty** — Never invent ETAs; show ranges when uncertain.
2. **Progressive commitment** — Browse lightly; intensify UI density only as user commits (cart → checkout).
3. **Recoverability** — Every error offers a next step.
4. **Thumb-first mobile** — Primary actions in lower 40% on phones.
5. **Glanceability for ops** — Courier/warehouse: status readable at arm’s length.
6. **Density with mercy** — Admin: high information density, but clear focal column.
7. **Accessibility by default** — WCAG 2.2 AA minimum on critical flows.
8. **Reduced motion respect** — Signature motions degrade gracefully.

## Product Principles

1. **Minutes are the product** — UI that slows discovery of ETA fails the brand.
2. **Assortment without overwhelm** — Prefer curated rails + strong search over infinite undifferentiated grids.
3. **Operational truth** — Customer UI never contradicts warehouse/courier state machines.
4. **One ecosystem language** — Same tokens; different density profiles.
5. **Flags over forks** — Experience variants via feature flags, not divergent design systems.
