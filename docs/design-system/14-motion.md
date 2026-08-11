# 14 — Motion Design

## Goals

Motion communicates **state change**, hierarchy, and live operations. Maximum **2–3 signature motions** per surface family. No ornamental parallax.

---

## Duration tokens

| Token | ms | Use |
|-------|-----|-----|
| `instant` | 0 | Reduced-motion hard cut |
| `fast` | 120 | Hover, color, opacity, tab fade |
| `micro` | 180 | Button press, icon toggles |
| `short` | 260 | Page push elements, check draw |
| `medium` | 320 | Sheets, cart commit (**signature**) |
| `long` | 440 | Complex multi-part |
| `xlong` | 600 | Rare emphasis |
| `cartCommit` | 320 | Add-to-cart morph |
| `orderPulse` | 400 | Status pulse once |
| `etaBreathe` | 1600 | Live ETA loop |

---

## Curves

| Token | Cubic-bezier | Use |
|-------|--------------|-----|
| `standard` | `0.2, 0.0, 0.0, 1.0` | Default |
| `emphasized` | `0.2, 0.0, 0.0, 1.0` | Enter emphasis (alias OK) |
| `decelerate` | `0.0, 0.0, 0.2, 1.0` | Decelerating entrances |
| `accelerate` | `0.3, 0.0, 1.0, 1.0` | Exits |

---

## Page transitions

See [13-navigation.md](13-navigation.md). Always honor reduced motion → opacity only or instant.

---

## Micro-interactions

| Control | Behavior |
|---------|----------|
| Button | Scale 0.98 on press `micro`; release spring to 1 |
| Switch | Thumb slide `fast` |
| Checkbox | Scale check-in `micro` |
| Favorite | Icon scale 0.8→1.1→1 `short` |
| Chip select | Background lerp `fast` |

---

## Loading animations

- `NxSpinner`: teal arc, 700ms linear loop
- Skeleton shimmer: 1200ms gradient sweep; pause if reduced motion (static bone)
- Pull-to-refresh: brand spinner

## Success animations

- Check draw-on `short`
- Payment success: illustration + single citrus spark (not confetti storm)
- Order placed: brief `orderPulse` then settle

## Checkout / cart

| Moment | Motion |
|--------|--------|
| Add to cart | Flying opacity optional; prefer local morph to qty + badge count tick |
| Cart bar appear | Slide up `medium` + elevation |
| Remove line | Height collapse `short` |
| Place order CTA | Loading replaces label; success navigates with medium fade |

## Map animations

- Camera ease `medium`–`long`
- Courier marker interpolation 1s easing between points
- Route polyline draw once; no continuous dash racing if distracting
- ETA breathe on chip only

## Gesture-driven

- Sheet dismiss velocity-aware
- Swipe delete reveal actions with rubber-band

## Hero

- Product image → PDP hero shared element when both tagged
- Disable across tabs; disable under reduced motion

## Signature set (customer)

1. **Cart commit** (medium morph)
2. **ETA breathe** (live)
3. **Order pulse** (status)

Courier/Warehouse/Admin: suppress 2–3; keep micro only.
