# 02 — Color System

> Token source: `tokens/nexora.tokens.json`  
> Runtime: `NxBrandColors`, `NxNeutralColors`, `NxColorRoles` in `nexora_design`

## Philosophy

Cool **graphite** neutrals + **teal ink** brand + sparse **citrus** signal. Color encodes status and hierarchy — never decoration for its own sake.

**Forbidden identity:** purple gradients, pastel candy palettes, terracotta-on-cream defaults.

---

## Primitive scales

### Brand — Teal (`primary`)

| Token | Hex | Use |
|-------|-----|-----|
| `brand.primary.50` | `#EBF8F8` | Soft brand wash, info surface light |
| `brand.primary.100` | `#D7F0F0` | Selected chip wash |
| `brand.primary.500` | `#0F8585` | Dark-mode brand text/link; hover lift |
| `brand.primary.600` | `#0B6E6E` | **Primary brand** — CTAs, focus, chrome |
| `brand.primary.700` | `#085858` | Pressed / textBrand light |
| `brand.primary.800` | `#064545` | Deep brand on large fills |

### Accent — Citrus (`accent`)

| Token | Hex | Use |
|-------|-----|-----|
| `brand.accent.100` | `#F7FAD4` | Soft accent wash |
| `brand.accent.400` | `#E8F07A` | **Signal** — primary CTA fill (customer), live pulse, dark nav active |
| `brand.accent.500` | `#D4DC5C` | Accent pressed |

Accent is rare: primary CTAs, live indicators, critical highlights. Never large background fields.

### Neutral — Graphite (`neutral`)

| Token | Hex |
|-------|-----|
| `n0` | `#FFFFFF` |
| `n25` | `#F7F8F8` |
| `n50` | `#F1F3F3` |
| `n100` | `#E6E9EA` |
| `n200` | `#CDD3D5` |
| `n300` | `#A8B1B4` |
| `n400` | `#7E898D` |
| `n500` | `#5C686C` |
| `n600` | `#3F4A4E` |
| `n700` | `#2A3336` |
| `n800` | `#1A2225` |
| `n900` | `#0B1214` |
| `n950` | `#070B0C` |

`n900` / ink `#0B1214` is brand ink for on-accent text.

### Semantic primitives

| Role | Light surface | Light fg | Dark surface | Dark fg |
|------|---------------|----------|--------------|---------|
| Success | `#E3F7EC` | `#1B7F4A` | `#143528` | `#3DDB8A` |
| Warning | `#FFF4E0` | `#B86E00` | `#3A2A0A` | `#FFC14D` |
| Danger | `#FDECEA` | `#C62828` | `#3A1414` | `#FF6B6B` |
| Info | `#EBF8F8` | `#0B6E6E` | `#0E2C2C` | `#5ED0D0` |

---

## Semantic roles (what apps consume)

Apps bind to **roles**, never primitives.

### Surfaces / background

| Role | Light | Dark | Purpose |
|------|-------|------|---------|
| `bg.canvas` | n25 | n950 | App scaffold behind content |
| `bg.surface` | n0 | n900 | Default panels / sheets content |
| `bg.surfaceRaised` | n0 | n800 | Cards / raised panels |
| `bg.sunken` | n50 | n950 | Wells, inset groups |
| `bg.nav` | n0 | n900 | Top/bottom nav chrome |
| `bg.brand` | primary600 | primary600 | Brand fills |
| `bg.accent` | accent400 | accent400 | Accent fills |
| `bg.disabled` | n100 | n800 | Disabled controls |
| `bg.overlay` | `#0B12147A` (48%) | `#00000099` (60%) | Modal scrim |

### Cards

Prefer `bg.surfaceRaised` + `border.subtle` + elevation level appropriate to density. Admin tables often skip card chrome — use sunken canvas + hairline rows.

### Borders

| Role | Light | Dark |
|------|-------|------|
| `border.subtle` | n100 | `#243034` |
| `border.default` | n200 | `#334044` |
| `border.strong` | n300 | n500 |
| `border.focus` | primary600 | primary500 |
| `border.danger` | danger600 | danger600 dark |

### Text hierarchy

| Role | Light | Dark | Use |
|------|-------|------|-----|
| `text.primary` | n700 | n50 | Body primary, titles |
| `text.secondary` | n500 | n300 | Supporting |
| `text.tertiary` | n400 | n400 | Meta, timestamps |
| `text.disabled` | n300 | n500 | Disabled |
| `text.inverse` | n0 | n900 | On dark/brand fills |
| `text.brand` | primary700 | primary500 | Brand emphasis text |
| `text.onBrand` | n0 | n0 | On teal fills |
| `text.onAccent` | ink n900 | ink n900 | On citrus fills |
| `text.link` | primary600 | primary500 | Links |

### Icons

| Role | Maps to |
|------|---------|
| `icon.primary` | text.primary |
| `icon.secondary` | text.secondary |
| `icon.brand` | brand primary |

### Navigation

| Role | Light | Dark |
|------|-------|------|
| `nav.itemDefault` | n500 | n400 |
| `nav.itemActive` | primary600 | accent400 |
| `nav.indicator` | primary600 | accent400 |

### Status (fg + surface pairs)

`success` / `successSurface`, `warning` / `warningSurface`, `danger` / `dangerSurface`, `info` / `infoSurface` — see semantic primitives.

---

## Light mode palette (summary)

- Canvas graphite wash; white surfaces
- Teal primary actions; citrus for highest-priority customer CTA
- Status greens/ambers/reds on soft tints
- Borders cool gray, never warm beige

## Dark mode palette (summary)

- Near-black canvas (`n950`); elevated `n800` for raised
- Prefer **surface step-up** over heavy shadows
- Active nav uses citrus for legibility on dark
- Status foregrounds brightened; surfaces deep tinted

---

## Contrast rules

| Pair | Minimum |
|------|---------|
| `text.primary` on `bg.surface` / `bg.canvas` | **4.5:1** (AA) |
| `text.secondary` on surface | **4.5:1** for essential UI; tertiary may be 3:1 only for non-essential |
| `text.onBrand` on `bg.brand` | **4.5:1** |
| `text.onAccent` on `bg.accent` | **4.5:1** (ink on citrus) |
| UI component non-text (icons, borders that convey state) | **3:1** |
| Focus ring vs adjacent | **3:1** |

### High contrast mode

- Force `text.primary` → near black/white extremes
- Borders → `border.strong`
- Disable low-contrast tertiary as sole information channel
- Accent CTAs gain 2px strong outline if needed

### Accessibility ratios — reference pairs (light)

| Foreground | Background | Approx ratio | Verdict |
|------------|------------|--------------|---------|
| n700 on n0 | — | ~12:1 | Pass AAA |
| n500 on n0 | — | ~5.5:1 | Pass AA |
| n0 on primary600 | — | ~5.2:1 | Pass AA |
| n900 on accent400 | — | ~8:1 | Pass AAA |
| danger600 on n0 | — | ~5.5:1 | Pass AA |

CI should fail token PRs that break AA on critical role pairs.

---

## Usage rules

1. Feature code: `NxTheme.of(context).colors.textPrimary` — never `Color(0xFF…)`.
2. Do not place citrus text on white for body copy — citrus is fill/signal only.
3. Danger is for irreversible / payment failure / stock-out critical — not decorative.
4. Maps: zone fills use `opacity.mapZone` (0.14) of brand/warning — not solid.
5. Disabled: `bg.disabled` + `text.disabled` + `opacity.disabled` (0.38) on icons if needed — do not rely on color alone.
