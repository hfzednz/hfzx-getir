# 05 — Border & Radius System

## Radius scale

| Token | px | Use |
|-------|-----|-----|
| `radius.none` | 0 | Tables, full-bleed media, admin dense grids |
| `radius.xs` | 4 | Chips micro, tags, tooltips |
| `radius.sm` | 8 | Inputs, dense cards, warehouse tiles, checkboxes |
| `radius.md` | 12 | **Default** — buttons, cards, sheets handles language |
| `radius.lg` | 16 | Bottom sheets top corners, large product cards |
| `radius.xl` | 24 | Marketing panels, rare hero containers |
| `radius.full` | 9999 | Avatars, circular FAB, true pills (use sparingly) |

**Brand default language = 12px (`md`).** Avoid rounded-full pill clusters.

---

## Border thickness

| Token | px | Use |
|-------|-----|-----|
| `border.hairline` | 1 | Default outlines, dividers |
| `border.thick` | 2 | Focus rings (offset), selected cards, segmented |
| `border.heavy` | 3 | High-contrast / error emphasis (rare) |

---

## Outline rules by component

| Component | Radius | Border | Notes |
|-----------|--------|--------|-------|
| Primary button (accent/brand fill) | md | none | Focus: 2px `border.focus` ring + 2px offset |
| Secondary button | md | hairline `border.default` | |
| Ghost / tertiary | md | none | |
| Icon button | md or full (circular) | none / hairline on secondary | |
| Text field | sm | hairline → thick on focus | Error → `border.danger` |
| Cards (customer product) | md–lg | hairline subtle OR elevation without border | Prefer one: border *or* shadow |
| Dialogs | lg | none + elevation 3 | |
| Bottom sheets | lg top / none bottom | none | Handle bar 32×4, radius full |
| Containers / sections | md | subtle optional | Admin often none |
| Chips / tags | xs–sm | optional | Selected: brand wash + border.focus |
| Checkboxes | xs | hairline | |
| Switches | full track | none | |

---

## Focus ring

- Width: 2px
- Color: `border.focus`
- Offset: 2px from control edge
- Always visible for keyboard; mouse may use subtle
- Never remove outline without equivalent
