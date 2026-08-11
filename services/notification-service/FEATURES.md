# Notification Service — Feature Matrix

| Area | Status | Notes |
|------|--------|-------|
| UpsertTemplate / Approve / Preview | ✅ | `{{var}}` render; versioned by key+channel+locale |
| RegisterDevice | ✅ | Push token upsert |
| SetPreferences / GetPreferences | ✅ | Channel opt-out + quiet hours |
| Send | ✅ | Prefs → render → enqueue → provider |
| SendBulk | ✅ | Per-principal idempotency suffix |
| ScheduleSend / Cancel / ProcessDue | ✅ | `send_at` delayed dispatch |
| MarkInboxRead / ListInbox | ✅ | In-app channel creates inbox item |
| GetDelivery / RetryFailed / MoveToDLQ | ✅ | Attempt history + DLQ after max retries |
| HandleDomainEvent | ✅ | Maps OrderDelivered etc. → template keys |
| BestSendTime / RecommendChannel | ✅ | AI stubs |
| AdminStats | ✅ | Status counts + DLQ size |
| Mock providers (FCM/APNs/SMTP/SMS/WA) | ✅ | Succeed by default |
| Memory repos | ✅ | Dev + tests |
| HTTP `:8101` `/v1/notifications/...` | ✅ | NEXORA error envelope |
| Outbox PublishPending | ✅ | Stub drain |
| PostgreSQL adapters | 🔶 | Migrations ready; repos stub |
| Kafka / Redis | 🔶 | Stub publisher/client |
| gRPC | 🔶 | Proto + stub server |

## Explicit non-goals

- CRM tickets
- Campaign / coupon ownership (`promotion-service`)
- Order aggregate ownership (`order-service`)

## Kafka topics

| Topic | Events |
|-------|--------|
| `notification.delivery` | NotificationQueued, NotificationSent, NotificationFailed, NotificationOpened, NotificationClicked |

## Inbound domain events (stub mapping)

OrderCreated, OrderDelivered, PaymentSuccess, RefundCompleted, RewardEarned, CouponIssued, MembershipUpgraded, SupportTicketCreated
