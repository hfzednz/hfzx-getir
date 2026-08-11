import 'package:flutter_test/flutter_test.dart';
import 'package:nexora_warehouse/features/picking/domain/entities/pick_task.dart';
import 'package:nexora_warehouse/shared/business_rules/picking_rules.dart';

void main() {
  const line = PickLine(
    id: 'l1',
    sku: 'SKU-1',
    barcode: '8690001',
    bin: 'A-01',
    qty: 3,
    pickedQty: 1,
    zone: 'A',
  );

  group('PickingRules.validateTransition', () {
    test('allows queued → claimed', () {
      final result = PickingRules.validateTransition(
        from: PickTaskStatus.queued,
        to: PickTaskStatus.claimed,
      );
      expect(result.isSuccess, isTrue);
    });

    test('blocks queued → staged', () {
      final result = PickingRules.validateTransition(
        from: PickTaskStatus.queued,
        to: PickTaskStatus.staged,
      );
      expect(result.isFailure, isTrue);
    });
  });

  group('PickingRules.validateLineScan', () {
    test('succeeds with matching barcode', () {
      final result = PickingRules.validateLineScan(
        status: PickTaskStatus.inProgress,
        line: line,
        scannedBarcode: '8690001',
        qtyDelta: 1,
      );
      expect(result.isSuccess, isTrue);
    });

    test('fails on barcode mismatch', () {
      final result = PickingRules.validateLineScan(
        status: PickTaskStatus.inProgress,
        line: line,
        scannedBarcode: 'wrong',
        qtyDelta: 1,
      );
      expect(result.isFailure, isTrue);
    });

    test('fails when over-picking', () {
      final result = PickingRules.validateLineScan(
        status: PickTaskStatus.inProgress,
        line: line,
        scannedBarcode: '8690001',
        qtyDelta: 5,
      );
      expect(result.isFailure, isTrue);
    });
  });

  group('PickingRules.validateShortPick', () {
    test('succeeds for remaining qty', () {
      final result = PickingRules.validateShortPick(
        status: PickTaskStatus.inProgress,
        line: line,
        missingQty: 2,
      );
      expect(result.isSuccess, isTrue);
    });

    test('fails when not in progress', () {
      final result = PickingRules.validateShortPick(
        status: PickTaskStatus.queued,
        line: line,
        missingQty: 1,
      );
      expect(result.isFailure, isTrue);
    });
  });

  group('PickingRules.canComplete', () {
    test('true when all lines complete', () {
      final task = PickTask(
        id: 't1',
        orderId: 'o1',
        status: PickTaskStatus.inProgress,
        lines: [
          PickLine(
            id: 'l1',
            sku: 's',
            barcode: 'b',
            bin: 'b1',
            qty: 1,
            pickedQty: 1,
          ),
        ],
      );
      expect(PickingRules.canComplete(task), isTrue);
    });
  });
}
