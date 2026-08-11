import 'package:nexora_core/nexora_core.dart';

import 'warehouse_role.dart';

/// Soft UI RBAC gates by [WarehouseRole] (BFF still enforces).
abstract final class RoleRules {
  static const picking = 'picking';
  static const packing = 'packing';
  static const dispatch = 'dispatch';
  static const inventory = 'inventory';
  static const transfers = 'transfers';
  static const purchasing = 'purchasing';
  static const quality = 'quality';
  static const returns = 'returns';
  static const reports = 'reports';
  static const ai = 'ai';
  static const map = 'map';
  static const overrides = 'overrides';
  static const multiStore = 'multi_store';

  static const Map<WarehouseRole, Set<String>> _access = {
    WarehouseRole.picker: {picking, map},
    WarehouseRole.packer: {packing, map},
    WarehouseRole.dispatcher: {dispatch, map},
    WarehouseRole.inventoryAuditor: {
      inventory,
      transfers,
      quality,
      returns,
      map,
    },
    WarehouseRole.supervisor: {
      picking,
      packing,
      dispatch,
      inventory,
      transfers,
      quality,
      returns,
      reports,
      ai,
      map,
      overrides,
    },
    WarehouseRole.warehouseManager: {
      picking,
      packing,
      dispatch,
      inventory,
      transfers,
      purchasing,
      quality,
      returns,
      reports,
      ai,
      map,
      overrides,
    },
    WarehouseRole.regionalManager: {
      inventory,
      transfers,
      purchasing,
      reports,
      ai,
      map,
      multiStore,
      overrides,
    },
    WarehouseRole.admin: {
      picking,
      packing,
      dispatch,
      inventory,
      transfers,
      purchasing,
      quality,
      returns,
      reports,
      ai,
      map,
      overrides,
      multiStore,
    },
  };

  static bool canAccess(WarehouseRole role, String module) {
    return _access[role]?.contains(module) ?? false;
  }

  static Set<String> modulesFor(WarehouseRole role) =>
      Set.unmodifiable(_access[role] ?? const {});

  static bool canPick(WarehouseRole role) => canAccess(role, picking);
  static bool canPack(WarehouseRole role) => canAccess(role, packing);
  static bool canDispatch(WarehouseRole role) => canAccess(role, dispatch);
  static bool canAdjustInventory(WarehouseRole role) =>
      canAccess(role, inventory);
  static bool canApproveTransfer(WarehouseRole role) =>
      canAccess(role, transfers) &&
      (role == WarehouseRole.warehouseManager ||
          role == WarehouseRole.regionalManager ||
          role == WarehouseRole.admin ||
          role == WarehouseRole.supervisor);
  static bool canOverride(WarehouseRole role) => canAccess(role, overrides);

  static Result<void> require(bool allowed, {required String action}) {
    if (allowed) return const Success(null);
    return Failure(
      NexoraValidationException(
        code: NexoraErrorCode.validationFailed,
        message: 'Role is not allowed to perform this action',
        details: {'action': action},
      ),
    );
  }

  static Result<void> requirePick(WarehouseRole role) =>
      require(canPick(role), action: 'pick');
}
