# Mobile Version Management

| App | pubspec | Tag pattern |
|-----|---------|-------------|
| Customer | `apps/mobile_customer/pubspec.yaml` | `mobile/customer/X.Y.Z+N` |
| Courier | `apps/mobile_courier/pubspec.yaml` | `mobile/courier/X.Y.Z+N` |
| Warehouse | `apps/mobile_warehouse/pubspec.yaml` | `mobile/warehouse/X.Y.Z+N` |

Rules:

1. Never reuse `+BUILD` / versionCode.
2. Hotfix increments PATCH + BUILD.
3. Backend compatibility via LiveOps min-version flags; force-update only for security.
4. Crash / ANR gates block production track promotion if crash-free < baseline − 0.5pp.
