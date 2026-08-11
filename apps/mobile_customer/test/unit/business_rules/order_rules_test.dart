import 'package:flutter_test/flutter_test.dart';
import 'package:nexora_customer/shared/business_rules/order_rules.dart';

void main() {
  group('OrderRules.isCancellable', () {
    test('allows cancel for confirmed order', () {
      expect(OrderRules.isCancellable(OrderLifecycleStatus.confirmed), isTrue);
    });

    test('blocks cancel for dispatched order', () {
      expect(OrderRules.isCancellable(OrderLifecycleStatus.dispatched), isFalse);
    });
  });

  group('OrderRules.validateCancel', () {
    test('succeeds for confirmed order', () {
      final result = OrderRules.validateCancel(
        status: OrderLifecycleStatus.confirmed,
        paymentCaptured: false,
      );
      expect(result.isSuccess, isTrue);
    });

    test('fails when payment is capturing', () {
      final result = OrderRules.validateCancel(
        status: OrderLifecycleStatus.pendingPayment,
        paymentCaptured: true,
      );
      expect(result.isFailure, isTrue);
    });
  });

  group('OrderRules.validatePartialCancel', () {
    test('succeeds during picking with subset of lines', () {
      final result = OrderRules.validatePartialCancel(
        status: OrderLifecycleStatus.picking,
        lineIdsToCancel: const ['l1'],
        totalLineCount: 3,
      );
      expect(result.isSuccess, isTrue);
    });

    test('fails when all lines selected', () {
      final result = OrderRules.validatePartialCancel(
        status: OrderLifecycleStatus.picking,
        lineIdsToCancel: const ['l1', 'l2'],
        totalLineCount: 2,
      );
      expect(result.isFailure, isTrue);
    });

    test('fails after dispatch', () {
      final result = OrderRules.validatePartialCancel(
        status: OrderLifecycleStatus.dispatched,
        lineIdsToCancel: const ['l1'],
        totalLineCount: 3,
      );
      expect(result.isFailure, isTrue);
    });
  });

  group('OrderRules.validateReorder', () {
    test('succeeds for delivered order with available items', () {
      final result = OrderRules.validateReorder(
        status: OrderLifecycleStatus.delivered,
        allItemsAvailable: true,
      );
      expect(result.isSuccess, isTrue);
    });

    test('fails when items unavailable', () {
      final result = OrderRules.validateReorder(
        status: OrderLifecycleStatus.delivered,
        allItemsAvailable: false,
      );
      expect(result.isFailure, isTrue);
    });

    test('fails for in-progress order', () {
      final result = OrderRules.validateReorder(
        status: OrderLifecycleStatus.picking,
        allItemsAvailable: true,
      );
      expect(result.isFailure, isTrue);
    });
  });
}
