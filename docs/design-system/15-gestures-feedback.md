# 15 — Gestures & Feedback

## Gestures

| Gesture | Where | Behavior |
|---------|-------|----------|
| Tap | Universal | Primary activate |
| Double tap | Media / map | Zoom or favorite (PDP image) — must be discoverable; never sole path |
| Long press | Qty, lists, map pin | Context menu / rapid qty; warehouse: continuous scan mode tips |
| Swipe horizontal | Cart item, tabs, banners | Reveal delete / dismiss |
| Swipe vertical | Sheets, maps | Dismiss / pan |
| Drag | Reorder (admin rare), map | |
| Pull to refresh | Lists, home, orders | Brand spinner; one-shot |
| Pinch | Maps, PDP gallery | Zoom |
| Map gestures | Tracking / address | Pan, pinch, two-finger rotate optional (prefer off for address refine) |
| Card gestures | Product rails | Horizontal scroll with peek; edge glow subtle |

### Rules

- Never require swipe as the only way to delete — also provide button
- Warehouse: prioritize scan button reliability over fancy gestures
- Conflict: vertical scroll wins over horizontal until horizontal intent clear (slop 8–12px)

---

## Haptics

| Event | iOS | Android |
|-------|-----|---------|
| Light tap / selection | lightImpact | CONTEXT_CLICK / light |
| Add to cart | medium | CONTEXT_CLICK |
| Success | success notification | CONFIRM |
| Error / reject | error | REJECT |
| SOS / critical | heavy | HEAVY |

Disable when system haptics off. No haptics on scroll.

---

## Visual feedback

| Kind | Spec |
|------|------|
| Pressed | Overlay `opacity.pressed` (0.12) or scale |
| Hover | Overlay `opacity.hover` (0.08) |
| Selection | Brand wash + border.focus |
| Focus | Focus ring (see borders) |
| Loading | Spinner / skeleton — disable double-submit |
| Error | border.danger + helper + optional shake once |
| Success | success surface flash or check |

Button feedback must complete within `micro` so UI feels instant even if network pending (optimistic where safe).
