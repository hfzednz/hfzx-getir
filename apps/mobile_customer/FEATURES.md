# Customer App — Business Features Status (Prompt 04)

> Extends Prompt 03. Architecture unchanged. Business rules live under `lib/shared/business_rules/`.

## Wired end-to-end

| Domain | Business behavior |
|--------|-------------------|
| Auth | OTP/email/social, refresh, devices, delete, export, privacy, biometric, **email verify**, **cart merge on login** |
| Addresses | Labels, map/GPS, building fields, favorite/default, zone validation |
| Home | Backend widget feed |
| Search | Instant/semantic, voice, image, barcode, **recent + trending** |
| Product | Variants, nutrition, favorites, rails, **share**, **price history**, **Q&A**, **bundles** |
| Favorites | product/brand/category/search + offline sync |
| Cart | Offline Drift+outbox, **coupon/gift/wallet/loyalty tenders**, inventory validate, `CartRules` |
| Checkout | Address/schedule/payment/review, **substitution + OOS + invoice + gift message**, **quote verify**, **PaymentRules retry**, installments/gift card |
| Orders | Cancel/partial/refund/reorder/receipt/invoice, **POD photos** |
| Tracking | WS refresh, map, **route polyline**, call, courier chat URL |
| Wallet / Loyalty / Coupons / Referral | Screens + **referral device fraud payload** |
| Support / Reviews | Tickets, AI assistant, FAQ; order/courier/**per-product** ratings |
| AI | **Hub + recipes** with priced add-to-cart |
| Analytics | Product view, cart add, checkout review, order placed/cancel |
| Offline | Cart + favorites + recent searches |

## Rule engines

- `CartRules` — min order, stock/qty, weight, age, coupon on cart, merge
- `CheckoutRules` + `CheckoutDraft` — address, schedule, invoice, gift, substitution conflict, final price tolerance
- `CouponRules` + selection helper — eligibility/stacking
- `OrderRules` — cancel / partial / reorder
- `PaymentRules` — idempotency, duplicate guard, retry cooldown

## Tests added (Prompt 04)

- `test/unit/business_rules/checkout_rules_test.dart`
- `test/unit/business_rules/payment_rules_test.dart`
- `test/unit/checkout_draft_test.dart`

Run: `flutter test test/unit/business_rules test/unit/checkout_draft_test.dart`
