# Supplier & Partner Ecosystem — Prompt #32

| Item | Value |
|------|--------|
| Service | `supplier-service` |
| HTTP | `:8117` `/v1/supplier` |
| Kafka | `supplier.events` |

## Non-ownership

- ERP COA/journals/AP 3-way match → `erp-service`
- Stock quantities → `inventory-service`
- Product PIM → `catalog-service`
- Payout execution → `settlement-service`
