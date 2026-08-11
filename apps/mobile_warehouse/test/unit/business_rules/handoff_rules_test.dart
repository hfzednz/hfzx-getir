import 'package:flutter_test/flutter_test.dart';
import 'package:nexora_warehouse/features/dispatch/domain/entities/handoff_task.dart';
import 'package:nexora_warehouse/shared/business_rules/handoff_rules.dart';

void main() {
  group('HandoffRules.validateHandoffScan', () {
    test('succeeds with matching token', () {
      final result = HandoffRules.validateHandoffScan(
        status: HandoffStatus.courierArrived,
        scannedToken: 'tok-1',
        expectedToken: 'tok-1',
        expectedOrderId: 'ord-1',
      );
      expect(result.isSuccess, isTrue);
    });

    test('fails on token mismatch', () {
      final result = HandoffRules.validateHandoffScan(
        status: HandoffStatus.courierArrived,
        scannedToken: 'wrong',
        expectedToken: 'tok-1',
        expectedOrderId: 'ord-1',
      );
      expect(result.isFailure, isTrue);
    });

    test('fails on order mismatch when provided', () {
      final result = HandoffRules.validateHandoffScan(
        status: HandoffStatus.verifying,
        scannedToken: 'tok-1',
        expectedToken: 'tok-1',
        expectedOrderId: 'ord-1',
        scannedOrderId: 'ord-2',
      );
      expect(result.isFailure, isTrue);
    });
  });

  group('HandoffRules.validateFailReason', () {
    test('accepts known reason', () {
      expect(
        HandoffRules.validateFailReason('qr_mismatch').isSuccess,
        isTrue,
      );
    });

    test('requires notes for other', () {
      expect(
        HandoffRules.validateFailReason('other').isFailure,
        isTrue,
      );
    });
  });

  group('HandoffRules.validateTransition', () {
    test('allows waiting → arrived', () {
      expect(
        HandoffRules.validateTransition(
          from: HandoffStatus.waitingCourier,
          to: HandoffStatus.courierArrived,
        ).isSuccess,
        isTrue,
      );
    });

    test('blocks waiting → handed off', () {
      expect(
        HandoffRules.validateTransition(
          from: HandoffStatus.waitingCourier,
          to: HandoffStatus.handedOff,
        ).isFailure,
        isTrue,
      );
    });
  });
}
