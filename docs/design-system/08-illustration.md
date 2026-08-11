# 08 — Illustration Style

## Philosophy

**Urban line + soft flat fills** in graphite/teal/citrus. Human silhouettes abstracted; products recognizable; no cartoon mascot dependency; no purple gradients.

Aspect: prefer simple center composition for empty states (max ~160–200px wide on mobile).

---

## Scenes (required set)

| Key | Narrative | Color notes |
|-----|-----------|-------------|
| `empty.generic` | Quiet shelf / open frame | neutrals + teal accent line |
| `empty.cart` | Open tote | citrus spark optional |
| `empty.search` | Magnifier + city grid | |
| `empty.orders` | Clipboard timeline | |
| `empty.favorites` | Outline heart shelf | |
| `empty.address` | Pin + block | |
| `success.generic` | Check in teal disc | |
| `success.order` | Door + bag | |
| `success.payment` | Shield + check | |
| `success.rewards` | Stamp / points | citrus |
| `error.generic` | Soft alert triangle | danger muted |
| `error.payment` | Card + alert | |
| `loading.brand` | Teal pulse mark (Lottie/Rive optional) | reduced-motion → static |
| `offline` | Signal slash + city | warning |
| `maintenance` | Cone + storefront | warning |
| `no_internet` | Same family as offline | |

---

## Copy pairing

Every illustration ships with:

1. Title (`title.md` / `headline.sm`)
2. One supporting sentence (`body.md` secondary)
3. One primary action (when recoverable)
4. Optional secondary text button

Voice: calm, direct — never blame the user.

---

## Technical

- SVG preferred; Lottie only for brand loading / success moments
- Dark mode: dedicated assets OR CSS/ColorFiltered carefully — test contrast
- Do not place text inside illustration files — text in UI
- Localization: no embedded words in artwork
