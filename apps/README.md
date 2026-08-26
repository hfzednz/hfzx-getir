# apps/

| App | Path | Stack | Dev port |
|-----|------|-------|----------|
| Customer web | `customer-web/` | Next.js + `@nexora/ui` | 3000 |
| Courier web | `courier-web/` | Next.js | 3001 |
| Warehouse web | `warehouse-web/` | Next.js | 3002 |
| Supplier web | `supplier-web/` | Next.js | 3003 |
| Finance web | `finance-web/` | Next.js | 3004 |
| Support web | `support-web/` | Next.js | 3005 |
| Operations web | `operations-web/` | Next.js | 3006 |
| Customer mobile | `mobile_customer/` | Flutter | — |
| Courier mobile | `mobile_courier/` | Flutter | — |
| Warehouse mobile | `mobile_warehouse/` | Flutter | — |
| Admin dashboard | `admin_web/` | Next.js + `@nexora/ui` | 3100 |
| Super Admin | `super_admin_web/` | Next.js + `@nexora/ui` | — |

Shared: `packages/web/core` (`@nexora/web-core`), `packages/web/ui` (`@nexora/ui`).

See `docs/implementation/FULL_WEB_PLATFORM.md`.
