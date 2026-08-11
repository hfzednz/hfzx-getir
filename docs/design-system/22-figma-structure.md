# 22 — Figma File Architecture

> Designers work in Figma; tokens sync from / to `tokens/nexora.tokens.json` via Tokens Studio or equivalent.

## File set

| File | Purpose |
|------|---------|
| `NEXORA — Foundations` | Tokens, type, grids, icon library |
| `NEXORA — Components` | Primitives + variants |
| `NEXORA — Patterns` | Product/order/search/map compositions |
| `NEXORA — Customer App` | Screens & flows |
| `NEXORA — Courier App` | Screens & flows |
| `NEXORA — Warehouse App` | Screens & flows |
| `NEXORA — Admin / Super Admin` | Dashboard IA & pages |
| `NEXORA — Illustrations` | Empty/success/error set |
| `NEXORA — Prototypes` | Motion / interactive |

---

## Pages (per Foundations file)

1. **Cover** — version, owners, links to constitution  
2. **Changelog**  
3. **Principles** — Kinetic Clarity summary  
4. **Color** — primitives + semantic role examples light/dark/HC  
5. **Typography** — scale specimens  
6. **Spacing / Grid** — 4pt, layout frames  
7. **Radius / Border / Elevation**  
8. **Motion** — duration board + GIFs/videos  
9. **Iconography** — component set  
10. **Accessibility** — contrast checkers, focus demos  

## Pages (Components file)

- Buttons, Inputs, Navigation, Feedback, Data Display, Overlays  
- Each as Component with Variant properties matching Flutter/web props  
- Dark mode as variant or separate theme mode via variables  

## Variables / Tokens

- Collections: `Color/Light`, `Color/Dark`, `Color/HighContrast`, `Spacing`, `Radius`, `Type`, `Motion`  
- Modes map 1:1 to JSON token groups  
- Publish library; apps consume as enabled library  

## Component property conventions

| Prop | Example values |
|------|----------------|
| `variant` | primary / secondary / tertiary / danger |
| `size` | sm / md / lg |
| `state` | default / hover / pressed / focus / disabled / loading |
| `density` | comfortable / compact / dense |

Naming: `Nx / Button` matching code `NxButton`.

## Assets

- `/icons/svg` exported  
- `/illustrations`  
- `/brand/logo` light/dark  

## Prototypes

- Customer: Home → Search → PDP → Cart → Checkout → Tracking  
- Courier: Offer → Navigate → Handoff  
- Warehouse: Claim → Scan → Pack  
- Admin: Live map → Order detail → Refund  

Use Smart Animate sparingly; document duration = tokens.

## Documentation page

- Do/Don't  
- Density matrix  
- Content voice examples  
- Handoff checklist to eng  

## Handoff rules

1. Specs link to token names, not only hex  
2. Redlines use spacing tokens  
3. Dev mode annotations include component name + variant  
4. Any new component requires Foundations/Components update **before** screen PR
