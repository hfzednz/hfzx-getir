# Promotion Service — Feature Matrix

| Area | Status | Notes |
|------|--------|-------|
| CreateCampaign | ✅ | draft / scheduled by starts_at |
| Activate / Pause / Expire | ✅ | Validated transitions |
| CreatePromotion + Rule | ✅ | percent\|fixed\|bogo\|bxgy\|bundle\|threshold\|free_ship\|gift\|multibuy |
| Rule priority / stack / exclusion | ✅ | ResolveConflicts |
| GenerateCoupon | ✅ | single\|multi\|personal\|public |
| RedeemCoupon | ✅ | Idempotent by key; single-use enforced |
| IssueVoucher / RedeemVoucher | ✅ | RemainingMinor balance |
| Evaluate (core) | ✅ | Active campaigns + schedule window + limits |
| Simulate | ✅ | Persists request/result |
| Admin list / overview | ✅ | Campaign list + counts |
| Usage counters | ✅ | global / user / order / device |
| Memory repos | ✅ | Dev + tests |
| HTTP `:8094` `/v1/promo/...` | ✅ | NEXORA error envelope, X-Tenant-Id |
| Outbox PublishPending | ✅ | Stub drain |
| Migrations | ✅ | campaigns→outbox + indexes |
| PostgreSQL adapters | 🔶 | Migrations ready; repos stub |
| Kafka / Redis | 🔶 | Stub publisher / active-campaign cache |
| gRPC | 🔶 | Proto + stub server |

## Explicit non-goals

- Payment PSP (`payment-service`)
- Loyalty points (`loyalty-service`)
- Order aggregate (`order-service`)
- Inventory / stock ledger
- Catalog product master (opaque ids only)

## Kafka topics

| Topic | Events |
|-------|--------|
| `promo.campaign` | CampaignCreated, CampaignActivated, CampaignPaused, CampaignExpired |
| `promo.coupon` | CouponGenerated, CouponRedeemed |
| `promo.voucher` | VoucherIssued, VoucherRedeemed |
| `promo.rule` | PromotionApplied |
