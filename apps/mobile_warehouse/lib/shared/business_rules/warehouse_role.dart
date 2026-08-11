/// Warehouse RBAC roles (ARCHITECTURE.md).
enum WarehouseRole {
  picker,
  packer,
  dispatcher,
  inventoryAuditor,
  supervisor,
  warehouseManager,
  regionalManager,
  admin;

  static WarehouseRole fromString(String? raw) {
    final v = (raw ?? '').toLowerCase().replaceAll('-', '_');
    return switch (v) {
      'picker' => WarehouseRole.picker,
      'packer' => WarehouseRole.packer,
      'dispatcher' => WarehouseRole.dispatcher,
      'inventory_auditor' || 'inventoryauditor' =>
        WarehouseRole.inventoryAuditor,
      'supervisor' => WarehouseRole.supervisor,
      'warehouse_manager' || 'warehousemanager' =>
        WarehouseRole.warehouseManager,
      'regional_manager' || 'regionalmanager' =>
        WarehouseRole.regionalManager,
      'admin' => WarehouseRole.admin,
      _ => WarehouseRole.picker,
    };
  }

  String get wireName => switch (this) {
        WarehouseRole.inventoryAuditor => 'inventory_auditor',
        WarehouseRole.warehouseManager => 'warehouse_manager',
        WarehouseRole.regionalManager => 'regional_manager',
        _ => name,
      };
}
