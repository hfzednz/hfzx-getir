# Loyalty Service — Feature Matrix

| Area | Status | Notes |
|------|--------|-------|
| EnsureAccount | ✅ | Opaque principal_id; default standard membership |
| EarnPoints / RedeemPoints | ✅ | Never negative; earn idempotent by order_id |
| Membership EvaluateUpgrade | ✅ | Threshold ladder (standard→silver→…→vip) |
| Rewards unlock/redeem | ✅ | Points cost deducted on redeem |
| Cashback grant → confirm | ✅ | pending→issued/failed; WalletClient.Credit |
| Referral Create/Apply/Complete | ✅ | Self-apply rejected; complete grants points |
| Missions track/complete | ✅ | Can unlock linked achievement |
| Streaks increment/break/recover | ✅ | Day-gap breaks |
| Spin weighted random | ✅ | Injectable Rand for tests |
| Achievements | ✅ | purchase_count, referral_count, spend_minor, mission_code |
| Collectibles | ✅ | Domain + repo (grant/list) |
| Leaderboard / AI scores | ✅ | Stubs |
| AdminManualGrant | ✅ | Audited |
| Memory repos | ✅ | Dev + tests |
| HTTP `:8093` `/v1/loyalty/...` | ✅ | NEXORA error envelope |
| Outbox PublishPending | ✅ | Stub drain |
| PostgreSQL adapters | 🔶 | Migrations ready; repos stub |
| Kafka / Redis / Wallet HTTP | 🔶 | Stub publishers/clients |
| gRPC | 🔶 | Proto + stub server |

## Explicit non-goals

- Payment PSP ownership (`payment-service`)
- Coupon/campaign engine (`promotion-service`)
- CRM tickets
- Wallet ledger ownership (`wallet-service` — Credit only for cashback/promo)

## Kafka topics

| Topic | Events |
|-------|--------|
| `loyalty.points` | PointsEarned, PointsRedeemed |
| `loyalty.rewards` | RewardUnlocked, RewardRedeemed |
| `loyalty.membership` | MembershipUpgraded, MembershipDowngraded |
| `loyalty.referral` | ReferralCompleted |
| `loyalty.cashback` | CashbackIssued |
| `loyalty.game` | AchievementUnlocked, MissionCompleted, StreakUpdated |
