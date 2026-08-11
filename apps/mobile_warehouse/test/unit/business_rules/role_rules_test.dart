import 'package:flutter_test/flutter_test.dart';
import 'package:nexora_warehouse/shared/business_rules/role_rules.dart';
import 'package:nexora_warehouse/shared/business_rules/warehouse_role.dart';

void main() {
  group('RoleRules capabilities', () {
    test('picker can pick only', () {
      expect(RoleRules.canPick(WarehouseRole.picker), isTrue);
      expect(RoleRules.canPack(WarehouseRole.picker), isFalse);
      expect(RoleRules.canDispatch(WarehouseRole.picker), isFalse);
    });

    test('packer can pack', () {
      expect(RoleRules.canPack(WarehouseRole.packer), isTrue);
      expect(RoleRules.canPick(WarehouseRole.packer), isFalse);
    });

    test('dispatcher can dispatch', () {
      expect(RoleRules.canDispatch(WarehouseRole.dispatcher), isTrue);
    });

    test('manager can approve transfer', () {
      expect(
        RoleRules.canApproveTransfer(WarehouseRole.warehouseManager),
        isTrue,
      );
      expect(
        RoleRules.canApproveTransfer(WarehouseRole.picker),
        isFalse,
      );
    });

    test('requirePick fails for packer', () {
      expect(
        RoleRules.requirePick(WarehouseRole.packer).isFailure,
        isTrue,
      );
    });

    test('admin can override', () {
      expect(RoleRules.canOverride(WarehouseRole.admin), isTrue);
    });
  });
}
