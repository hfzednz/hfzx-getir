import 'package:flutter_test/flutter_test.dart';
import 'package:nexora_warehouse/shared/business_rules/inventory_rules.dart';

void main() {
  group('InventoryRules.validateAdjust', () {
    test('accepts valid adjustment', () {
      final result = InventoryRules.validateAdjust(
        delta: -2,
        reasonCode: 'damage',
        currentOnHand: 10,
      );
      expect(result.isSuccess, isTrue);
    });

    test('rejects zero delta', () {
      final result = InventoryRules.validateAdjust(
        delta: 0,
        reasonCode: 'damage',
        currentOnHand: 10,
      );
      expect(result.isFailure, isTrue);
    });

    test('rejects negative resulting stock', () {
      final result = InventoryRules.validateAdjust(
        delta: -20,
        reasonCode: 'damage',
        currentOnHand: 5,
      );
      expect(result.isFailure, isTrue);
    });

    test('requires notes for other', () {
      final result = InventoryRules.validateAdjust(
        delta: 1,
        reasonCode: 'other',
        currentOnHand: 5,
      );
      expect(result.isFailure, isTrue);
    });
  });

  group('InventoryRules stock helpers', () {
    test('low and oos flags', () {
      expect(InventoryRules.isOutOfStock(0), isTrue);
      expect(
        InventoryRules.isLowStock(onHand: 2, reorderPoint: 5),
        isTrue,
      );
      expect(
        InventoryRules.isLowStock(onHand: 0, reorderPoint: 5),
        isFalse,
      );
    });

    test('FEFO preference', () {
      final early = DateTime(2026, 1, 1);
      final late = DateTime(2026, 6, 1);
      expect(
        InventoryRules.preferFefo(
          candidateExpiry: early,
          otherExpiry: late,
        ),
        isTrue,
      );
    });
  });

  group('InventoryRules.validateCycleCountQty', () {
    test('rejects negative', () {
      expect(
        InventoryRules.validateCycleCountQty(countedQty: -1).isFailure,
        isTrue,
      );
    });
  });
}
