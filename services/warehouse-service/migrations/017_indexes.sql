-- Performance indexes for warehouse ops read/write paths.

-- Warehouses
CREATE INDEX idx_warehouses_tenant_status ON warehouses (tenant_id, status)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_warehouses_tenant_type ON warehouses (tenant_id, type)
    WHERE deleted_at IS NULL;

-- Stations
CREATE INDEX idx_stations_warehouse_type ON stations (warehouse_id, type)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_stations_warehouse_status ON stations (warehouse_id, status)
    WHERE deleted_at IS NULL;

-- Fulfillment orders
CREATE INDEX idx_fulfillment_orders_wh_status ON fulfillment_orders (warehouse_id, status, created_at DESC);
CREATE INDEX idx_fulfillment_orders_tenant_status ON fulfillment_orders (tenant_id, status, created_at DESC);
CREATE INDEX idx_fulfillment_orders_sla ON fulfillment_orders (warehouse_id, sla_deadline)
    WHERE status NOT IN ('dispatched', 'cancelled', 'failed') AND sla_deadline IS NOT NULL;
CREATE INDEX idx_fulfillment_orders_priority ON fulfillment_orders (warehouse_id, priority DESC, created_at)
    WHERE status IN ('received', 'reserved', 'pick_queued', 'pack_queued', 'dispatch_queued');
CREATE INDEX idx_fulfillment_orders_reservation ON fulfillment_orders (reservation_id)
    WHERE reservation_id IS NOT NULL;
CREATE INDEX idx_fulfillment_orders_express ON fulfillment_orders (warehouse_id, created_at)
    WHERE express = TRUE AND status NOT IN ('dispatched', 'cancelled', 'failed');

-- Fulfillment lines
CREATE INDEX idx_fulfillment_lines_fulfillment ON fulfillment_lines (fulfillment_id, sort_order);
CREATE INDEX idx_fulfillment_lines_variant ON fulfillment_lines (variant_id);
CREATE INDEX idx_fulfillment_lines_location ON fulfillment_lines (location_code)
    WHERE location_code <> '';

-- Tasks — claim queue
CREATE INDEX idx_tasks_wh_status_priority ON tasks (warehouse_id, status, priority DESC, created_at)
    WHERE status IN ('queued', 'claimed', 'in_progress', 'escalated');
CREATE INDEX idx_tasks_wh_type_queued ON tasks (warehouse_id, type, priority DESC, created_at)
    WHERE status = 'queued';
CREATE INDEX idx_tasks_assignee ON tasks (assignee_id, status)
    WHERE assignee_id IS NOT NULL;
CREATE INDEX idx_tasks_fulfillment ON tasks (fulfillment_id)
    WHERE fulfillment_id IS NOT NULL;
CREATE INDEX idx_tasks_wave ON tasks (wave_id)
    WHERE wave_id IS NOT NULL;
CREATE INDEX idx_tasks_batch ON tasks (batch_id)
    WHERE batch_id IS NOT NULL;
CREATE INDEX idx_tasks_sla ON tasks (warehouse_id, sla_deadline)
    WHERE status IN ('queued', 'claimed', 'in_progress') AND sla_deadline IS NOT NULL;

-- Pick sessions / scans
CREATE INDEX idx_pick_sessions_warehouse ON pick_sessions (warehouse_id, created_at DESC);
CREATE INDEX idx_pick_sessions_fulfillment ON pick_sessions (fulfillment_id)
    WHERE fulfillment_id IS NOT NULL;
CREATE INDEX idx_pick_scans_session ON pick_scans (session_id, scanned_at);
CREATE INDEX idx_pick_scans_line ON pick_scans (line_id, scanned_at);
CREATE INDEX idx_pick_scans_fail ON pick_scans (session_id)
    WHERE ok = FALSE;

-- Pack sessions
CREATE INDEX idx_pack_sessions_station ON pack_sessions (station_id, created_at DESC);
CREATE INDEX idx_pack_sessions_fulfillment ON pack_sessions (fulfillment_id)
    WHERE fulfillment_id IS NOT NULL;

-- Dispatch units
CREATE INDEX idx_dispatch_units_wh_status ON dispatch_units (warehouse_id, status, created_at);
CREATE INDEX idx_dispatch_units_fulfillment ON dispatch_units (fulfillment_id);
CREATE INDEX idx_dispatch_units_courier ON dispatch_units (warehouse_id, courier_ref)
    WHERE courier_ref <> '';

-- QC
CREATE INDEX idx_qc_inspections_wh_result ON qc_inspections (warehouse_id, result, created_at DESC);
CREATE INDEX idx_qc_inspections_fulfillment ON qc_inspections (fulfillment_id)
    WHERE fulfillment_id IS NOT NULL;

-- Workforce
CREATE INDEX idx_employees_warehouse_status ON employees (warehouse_id, status)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_employees_principal ON employees (principal_id)
    WHERE principal_id IS NOT NULL;
CREATE INDEX idx_shifts_warehouse_status ON shifts (warehouse_id, status, planned_start);
CREATE INDEX idx_shifts_employee ON shifts (employee_id, planned_start DESC);
CREATE INDEX idx_attendance_employee_time ON attendance_events (employee_id, occurred_at DESC);
CREATE INDEX idx_attendance_shift ON attendance_events (shift_id, occurred_at)
    WHERE shift_id IS NOT NULL;
CREATE INDEX idx_break_periods_shift ON break_periods (shift_id, started_at);

-- Equipment / sensors
CREATE INDEX idx_equipment_warehouse_kind ON equipment (warehouse_id, kind)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_equipment_status ON equipment (warehouse_id, status)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_equipment_heartbeats_eq ON equipment_heartbeats (equipment_id, recorded_at DESC);
CREATE INDEX idx_sensor_readings_wh_time ON sensor_readings (warehouse_id, recorded_at DESC);
CREATE INDEX idx_sensor_readings_metric ON sensor_readings (warehouse_id, metric, recorded_at DESC);
CREATE INDEX idx_sensor_readings_equipment ON sensor_readings (equipment_id, recorded_at DESC)
    WHERE equipment_id IS NOT NULL;

-- Labels / messages
CREATE INDEX idx_labels_wh_kind ON labels (warehouse_id, kind, created_at DESC);
CREATE INDEX idx_labels_fulfillment ON labels (fulfillment_id)
    WHERE fulfillment_id IS NOT NULL;
CREATE INDEX idx_labels_status ON labels (warehouse_id, status)
    WHERE status IN ('draft', 'ready');
CREATE INDEX idx_messages_wh_time ON messages (warehouse_id, created_at DESC);
CREATE INDEX idx_messages_tenant_time ON messages (tenant_id, created_at DESC);
CREATE INDEX idx_messages_active ON messages (warehouse_id, severity, created_at DESC)
    WHERE expires_at IS NULL OR expires_at > now();

-- Audit
CREATE INDEX idx_warehouse_audit_tenant_time ON warehouse_audit_events (tenant_id, created_at DESC);
CREATE INDEX idx_warehouse_audit_resource ON warehouse_audit_events (resource_type, resource_id);
CREATE INDEX idx_warehouse_audit_warehouse ON warehouse_audit_events (warehouse_id, created_at DESC)
    WHERE warehouse_id IS NOT NULL;
CREATE INDEX idx_warehouse_audit_fulfillment ON warehouse_audit_events (fulfillment_id, created_at DESC)
    WHERE fulfillment_id IS NOT NULL;
CREATE INDEX idx_warehouse_audit_task ON warehouse_audit_events (task_id, created_at DESC)
    WHERE task_id IS NOT NULL;
