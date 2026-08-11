# 20 — Accessibility

## Standard

**WCAG 2.2 Level AA** minimum on customer critical journeys (auth, browse, cart, checkout, tracking, support). Strive AAA for text contrast on primary reading surfaces where feasible.

Ops apps: AA for text; outdoor glare → offer high-contrast theme.

---

## Screen readers

- All interactive elements have names
- Images: product name as alt; decorative empty
- Live regions: ETA updates, order status, cart count (`polite`); payment errors (`assertive`)
- Custom controls expose roles (checkbox, tab, switch)
- Announce sheet open/close

## Dynamic text

- Support 100–200% system scaling
- No clipped CTAs at 135%
- Prefer reflow over horizontal scroll for text

## High contrast

- Theme mode `highContrast` intensifies borders and text roles
- Charts use patterns + color

## Reduced motion

- Honor OS `prefers-reduced-motion` / Flutter `disableAnimations`
- Replace loops with static; heroes → fade; shimmer → static skeleton
- Never convey info by motion alone

## Keyboard (web / admin)

- Tab order logical
- Skip link to main content
- Focus visible always
- Esc closes dialogs/sheets/menus
- Arrow keys in menus, listboxes, tabs
- Command palette fully keyboard operable

## Focus management

- On route change: focus main heading or first field
- On dialog open: focus first control; trap focus; restore on close
- On async refresh: do not steal focus unless user submitted

## Touch targets

- Minimum 44×44 CSS/dp (48 preferred customer)
- Spacing ≥ 8 between adjacent targets

## Color

- Never status by color alone — icon + text
- Link underline or weight difference in body copy

## Testing gates

- Flutter: semantics debugger checks on critical flows
- Web: axe-core on Storybook
- Manual VoiceOver / TalkBack pass per release train for checkout & tracking
