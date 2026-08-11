# Supplier / procurement collaboration runbook

## Supplier stuck pending

1. Verify documents uploaded (`POST .../documents`).
2. `POST /v1/supplier/suppliers/{id}/approve`.
3. Confirm `SupplierApproved` on `supplier.events` and `erpSupplierId` set.

## Invoice match signal vs ERP AP

Match signals here are collaboration hints. Posting AP / true 3-way match remains in `erp-service`.
