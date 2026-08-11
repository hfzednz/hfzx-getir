# 11 — Order Components

## Cart Item — `NxCartItem`

Horizontal: image 64–72, name, variant, price, `NxQtySelector`, remove, substitution preference link. Swipe-to-delete optional (confirm). Unavailable state with replace CTA.

## Checkout Summary — `NxCheckoutSummary`

Lines: subtotal, discounts, delivery fee, tip, tax display (city rules), **total** emphasized. Fee honesty: show “Free delivery” only when truly zero. Expandable details.

## ETA Card — `NxEtaCard`

**Signature component.** Large tabular ETA; range or single; live pulse on accent/brand dot; secondary line “Usually 8–14 min”. Reduced motion: static.

## Courier Card — `NxCourierCard`

Avatar, name, vehicle, rating optional, call/chat actions; privacy-safe (masked phone via dialer).

## Tracking Timeline — `NxTrackingTimeline`

Vertical `NxTimeline`: Placed → Picking → On the way → Delivered. Current step emphasized; timestamps caption; failure branch in danger.

## Delivery Progress — `NxDeliveryProgress`

Map + compact status bar combo; progress not fake — bind to real state machine.

## Order Status — `NxOrderStatusChip`

Mapped colors:

| Status | Tone |
|--------|------|
| Placed / Paid | info |
| Picking / Packed | brand |
| Assigned / Picked up / In transit | accent/live |
| Delivered | success |
| Cancelled / Failed | danger |
| Delayed | warning |

## Receipt — `NxReceipt`

Printable structure; mono for codes; tabular money; share/download actions.

## Invoice — `NxInvoiceView`

Admin/customer legal layout; dense table; brand header restrained; download PDF affordance.

---

## Composition rules

1. Tracking screen first viewport: ETA + map + next status — not promo clutter  
2. Cart bar (`NxCartBar`): item count, total, CTA “View cart” / “Checkout”  
3. Never show courier location without active in-transit state  
4. Cancel window: clear countdown if time-boxed
