# 10 — Product Components

## Product Card family — `NxProductCard`

Shared data slots: image, name (2-line clamp), price block, unit, discount badge, favorite, qty / add, stock, delivery ETA chip (optional), promo tag.

### Variants

| Variant | Layout | Use |
|---------|--------|-----|
| `grid` | Vertical; image 1:1; content below | Home rails, category grid |
| `large` | Wider image; richer price / promo | Featured merchandising |
| `compact` | Smaller image 72–88; denser type | Search results dense |
| `horizontal` | Image left 88–112; content right | Search list, reorder |
| `list` | Full-width row; minimal chrome | Warehouse-adjacent customer rare; admin catalog |

### Anatomy (grid)

```
┌─────────────────┐
│  [♡]     [-30%] │  favorite top-end; discount top-start
│     IMAGE       │  radius.md top; out-of-stock opacity.imageOutOfStock
│                 │
├─────────────────┤
│ Name            │  title.sm, 2 lines
│ Unit · Meta     │  caption
│ ₺Price  ₺Was    │  NxPriceBlock
│ [  −  1  +  ]   │  OR [ Add ] → morphs to qty
└─────────────────┘
```

Radius: md–lg. Border subtle OR elevation 1 — not both heavy.

### Interaction

- Tap card → PDP
- Favorite toggles without navigation (optimistic)
- Add: `NxQtySelector` appears with cart-commit motion
- Out of stock: disabled add; show substitute CTA if available
- Image lazy + blurhash/LQIP

---

## Product Detail — `NxProductDetailHeader`

- Full-bleed image gallery (edge-to-edge on mobile — constitution)
- Sticky collapse title + price on scroll
- Sections one job each: info, nutrition, reviews, similar
- Primary CTA sticky bottom with qty + add

## Quantity Selector — `NxQtySelector`

Heights 32/40; minus / value / plus; long-press repeat; haptics light; min 0 removes line in cart context; max from stock.

## Discount Badge — `NxDiscountBadge`

caption.sm bold; danger or brand surface; formats `%` or absolute per locale.

## Favorite Button — `NxFavoriteButton`

Icon button; outlined → filled; no sticker burst.

## Stock Indicator — `NxStockIndicator`

| State | Visual |
|-------|--------|
| In stock | Hidden or subtle |
| Low | warning text “Only N left” |
| Out | danger / tertiary “Out of stock” |

## Nutrition Card — `NxNutritionCard`

Accordion; table.cell tabular for values; allergens emphasized.

## Review Card — `NxReviewCard`

Avatar, rating, date, body clamp, merchant reply.

## Price Block — `NxPriceBlock`

| Slot | Style |
|------|-------|
| Current | price.md/lg tabular |
| Was | caption strikethrough tertiary |
| Unit price | caption secondary |

Currency from locale; amount from minor units formatted upstream.

## Promotion Banner — `NxPromoBanner`

Horizontal rail card; image optional; **not** overlaid on hero as sticker.

## Delivery Badge — `NxDeliveryBadge`

ETA chip: icon + `eta.md` or caption; uses time honesty — ranges when uncertain.
