# 16 — Loading System

## Principles

1. Prefer **skeleton** mirroring final layout over spinners for pages  
2. Keep previous content visible during refresh when possible  
3. Always offer **retry** on failure  
4. Offline is a first-class state, not a blank error  

---

## Skeleton UI — `NxSkeleton`

- Bone color: `bg.sunken` with shimmer highlight toward `bg.surface`
- Radius matches target component (image md, text xs)
- Product grid: 2–3 placeholder cards
- Admin table: row bones

## Shimmer

- Duration 1200ms linear sweep
- Reduced motion: static bones, no sweep

## Progressive loading

| Priority | Content |
|----------|---------|
| P0 | App chrome, address/ETA chip |
| P1 | Above-fold rails / primary task |
| P2 | Secondary rails, reviews |
| P3 | Recommendations |

Images: LQIP/blurhash → full; decode off main thread.

## Lazy loading / infinite scroll

- Prefetch when 2 screens from end
- Footer loader `NxSpinner` sm
- End-of-list caption — no infinite spinner trap
- Preserve scroll offset on back

## Offline states — `NxOfflineBanner` + empty

- Banner sticky under top bar when connectivity lost
- Cached catalog readable; checkout blocked with clear reason
- Courier/Warehouse: outbox pending count chip

## Retry states

- Inline `NxErrorState` with Retry button
- Exponential backoff for auto-retry on tracking only (cap announced)

## Blocking vs non-blocking

| Case | Pattern |
|------|---------|
| Initial route load | Full skeleton |
| Button action | Button loading |
| Background sync | Subtle top/offline chip |
| Payment | Blocking overlay with cancel policy documented |
