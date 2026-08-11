# 12 — Flutter & Web UI Architecture

> Spec only — no business logic.  
> **Mandatory companion:** [00-INDEX.md](00-INDEX.md) and all DS docs.  
> Tokens SoT: [`tokens/nexora.tokens.json`](tokens/nexora.tokens.json) → codegen into `nexora_design` / `@nexora/ui`.

## Flutter — `nexora_design`

### Widget organization

```text
packages/flutter/nexora_design/
  lib/
    nexora_design.dart              # barrel
    src/
      tokens/                       # generated from JSON
        colors.dart
        typography.dart
        spacing.dart
        radius.dart
        elevation.dart
        motion.dart
        breakpoints.dart
        opacity.dart
      theme/
        nx_theme.dart               # ThemeExtension
        nx_theme_data.dart
        brightness_modes.dart
        density.dart                # comfortable|compact|dense
      foundations/
        nx_text.dart
        nx_icon.dart
        nx_scaffold.dart
      components/
        button/
        field/
        sheet/
        …                           # one folder per component
      patterns/
        product_card/
        order_timeline/
        cart_bar/
        eta_card/
        map_chrome/
        …
      icons/
      illustrations/
      a11y/
        focus_helpers.dart
        live_region.dart
```

### Theme architecture

- `ThemeData` + `NxThemeExtension` holding semantic colors, motion, density
- Apps select `NxTheme.light()` / `NxTheme.dark()` / high contrast
- Feature code reads `NxTheme.of(context).colors.textPrimary` — never raw `Color(0xFF…)`
- Fonts registered in app bootstrap (Satoshi, Geist, Geist Mono)

### Component architecture

- Presentational only; callbacks in, no repositories
- `NxButton` etc. use tokens exclusively
- Golden tests per component × variant × theme
- Widgetbook catalog in `apps/widgetbook` (or package gallery)

### Responsive / adaptive

- `NxBreakpoint` helpers from MediaQuery
- `NxAdaptiveScaffold` switches BottomNav / Rail / Side
- Density from app flavor: Customer comfortable; Courier/Warehouse compact

### Design token implementation

1. Edit `docs/design-system/tokens/nexora.tokens.json`
2. Codegen → Dart + CSS + TS
3. PR fails if Figma export hash mismatches (CI phase)

### App usage

```dart
// illustrative contract — not production code in this prompt
MaterialApp(
  theme: NxTheme.light(density: NxDensity.comfortable),
  darkTheme: NxTheme.dark(density: NxDensity.comfortable),
  home: const CustomerRoot(),
);
```

---

## Web — `@nexora/ui`

### Structure

```text
packages/web/ui/
  src/
    tokens/         # CSS variables + TS const
    theme/
    components/     # NxButton.tsx …
    patterns/
    charts/
    styles.css
```

### Rules

- CSS variables from `nexora.css`
- Radix/Headless primitives OK **if** restyled to NEXORA tokens (no default look leaking)
- TanStack Table skin via `NxDataTable`
- Storybook mandatory

### Admin layout shell

`NxAdminShell`: sidebar + city switcher + top bar + command palette region + content canvas.

---

## Cross-platform parity matrix

| Token domain | Flutter | Web |
|--------------|---------|-----|
| Color semantic | ThemeExtension | CSS vars |
| Type | TextStyle tokens | utility classes / style objects |
| Space | EdgeInsets tokens | spacing scale |
| Motion | AnimationCurves/Durations | CSS / Motion One / Framer |
| Icons | NxIcon | SVG sprite / React icon |

Parity is visual and naming—not pixel-identical Material clones.

---

## Quality gates

- No raw Material `ElevatedButton` in apps—wrap or replace with `NxButton`
- Contrast CI on token pairs
- Axe checks on Storybook for web components
- Golden diffs for Flutter on PR
