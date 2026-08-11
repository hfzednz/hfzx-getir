-- Performance indexes for inventory read/write and lookup paths.

-- Warehouses
CREATE INDEX idx_warehouses_tenant_status ON warehouses (tenant_id, status)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_warehouses_tenant_region ON warehouses (tenant_id, region_id)
    WHERE deleted_at IS NULL AND region_id IS NOT NULL;

-- Locations
CREATE INDEX idx_locations_warehouse_parent ON locations (warehouse_id, parent_id)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_locations_warehouse_kind ON locations (warehouse_id, kind)
    WHERE deleted_at IS NULL AND is_active = TRUE;
CREATE INDEX idx_locations_path ON locations (warehouse_id, path text_pattern_ops)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_locations_zone_type ON locations (warehouse_id, zone_type)
    WHERE deleted_at IS NULL AND zone_type IS NOT NULL;

-- Stock balances — critical (variant_id, warehouse_id) lookup
CREATE INDEX idx_stock_balances_variant_warehouse ON stock_balances (variant_id, warehouse_id);
CREATE INDEX idx_stock_balances_warehouse_variant ON stock_balances (warehouse_id, variant_id);
CREATE INDEX idx_stock_balances_tenant_warehouse ON stock_balances (tenant_id, warehouse_id);
CREATE INDEX idx_stock_balances_sku ON stock_balances (tenant_id, sku_code)
    WHERE sku_code <> '';
CREATE INDEX idx_stock_balances_location ON stock_balances (location_id)
    WHERE location_id IS NOT NULL;
CREATE INDEX idx_stock_balances_reorder ON stock_balances (warehouse_id)
    WHERE on_hand <= reorder_point;

-- Lots — expiry / FEFO
CREATE INDEX idx_stock_lots_expiry ON stock_lots (warehouse_id, expiry_date)
    WHERE status = 'good' AND expiry_date IS NOT NULL;
CREATE INDEX idx_stock_lots_balance ON stock_lots (balance_id)
    WHERE status = 'good';
CREATE INDEX idx_stock_lots_variant_expiry ON stock_lots (variant_id, warehouse_id, expiry_date)
    WHERE status = 'good';
CREATE INDEX idx_stock_lots_near_expiry ON stock_lots (expiry_date)
    WHERE status = 'good' AND expiry_date IS NOT NULL AND qty > 0;

-- Reservations — expires_at sweep
CREATE INDEX idx_reservations_expires_at ON reservations (expires_at)
    WHERE status = 'active' AND expires_at IS NOT NULL;
CREATE INDEX idx_reservations_tenant_status ON reservations (tenant_id, status, created_at DESC);
CREATE INDEX idx_reservations_external_ref ON reservations (tenant_id, external_ref)
    WHERE external_ref <> '';
CREATE INDEX idx_reservations_warehouse ON reservations (warehouse_id)
    WHERE warehouse_id IS NOT NULL AND status = 'active';

-- Reservation lines
CREATE INDEX idx_reservation_lines_reservation ON reservation_lines (reservation_id);
CREATE INDEX idx_reservation_lines_variant_wh ON reservation_lines (variant_id, warehouse_id);
CREATE INDEX idx_reservation_lines_balance ON reservation_lines (balance_id)
    WHERE balance_id IS NOT NULL;
CREATE INDEX idx_reservation_lines_lot ON reservation_lines (lot_id)
    WHERE lot_id IS NOT NULL;

-- Movements
CREATE INDEX idx_stock_movements_balance_time ON stock_movements (balance_id, occurred_at DESC)
    WHERE balance_id IS NOT NULL;
CREATE INDEX idx_stock_movements_variant_wh ON stock_movements (variant_id, warehouse_id, occurred_at DESC);
CREATE INDEX idx_stock_movements_tenant_time ON stock_movements (tenant_id, occurred_at DESC);
CREATE INDEX idx_stock_movements_type ON stock_movements (warehouse_id, type, occurred_at DESC);
CREATE INDEX idx_stock_movements_reservation ON stock_movements (reservation_id)
    WHERE reservation_id IS NOT NULL;

-- Transfers
CREATE INDEX idx_transfers_tenant_status ON transfers (tenant_id, status, created_at DESC);
CREATE INDEX idx_transfers_from_wh ON transfers (from_warehouse_id, status);
CREATE INDEX idx_transfers_to_wh ON transfers (to_warehouse_id, status);
CREATE INDEX idx_transfer_lines_transfer ON transfer_lines (transfer_id);
CREATE INDEX idx_transfer_lines_variant ON transfer_lines (variant_id);

-- Counts
CREATE INDEX idx_count_sessions_wh_status ON count_sessions (warehouse_id, status, created_at DESC);
CREATE INDEX idx_count_sessions_tenant ON count_sessions (tenant_id, status);
CREATE INDEX idx_count_lines_session ON count_lines (session_id);
CREATE INDEX idx_count_lines_variant ON count_lines (variant_id);
CREATE INDEX idx_count_lines_variance ON count_lines (session_id)
    WHERE variance IS NOT NULL AND variance <> 0;

-- Returns
CREATE INDEX idx_inventory_returns_wh_status ON inventory_returns (warehouse_id, status, created_at DESC);
CREATE INDEX idx_inventory_returns_external ON inventory_returns (tenant_id, external_ref)
    WHERE external_ref <> '';
CREATE INDEX idx_inventory_return_lines_return ON inventory_return_lines (return_id);
CREATE INDEX idx_inventory_return_lines_variant ON inventory_return_lines (variant_id);

-- Purchase receipts
CREATE INDEX idx_purchase_receipts_wh_status ON purchase_receipts (warehouse_id, status, created_at DESC);
CREATE INDEX idx_purchase_receipts_po_ref ON purchase_receipts (tenant_id, po_ref);
CREATE INDEX idx_purchase_receipt_lines_receipt ON purchase_receipt_lines (receipt_id);
CREATE INDEX idx_purchase_receipt_lines_variant ON purchase_receipt_lines (variant_id);

-- Forecasts
CREATE INDEX idx_stock_forecasts_variant_wh ON stock_forecasts (variant_id, warehouse_id, horizon_start);
CREATE INDEX idx_stock_forecasts_wh_horizon ON stock_forecasts (warehouse_id, horizon_start, horizon_end);

-- Audit
CREATE INDEX idx_inventory_audit_tenant_time ON inventory_audit_events (tenant_id, created_at DESC);
CREATE INDEX idx_inventory_audit_resource ON inventory_audit_events (resource_type, resource_id);
CREATE INDEX idx_inventory_audit_warehouse ON inventory_audit_events (warehouse_id, created_at DESC)
    WHERE warehouse_id IS NOT NULL;
CREATE INDEX idx_inventory_audit_variant ON inventory_audit_events (variant_id, created_at DESC)
    WHERE variant_id IS NOT NULL;
