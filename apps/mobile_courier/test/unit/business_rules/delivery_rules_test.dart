import 'package:flutter_test/flutter_test.dart';
import 'package:nexora_courier/features/deliveries/domain/entities/delivery_job.dart';
import 'package:nexora_courier/shared/business_rules/delivery_rules.dart';

void main() {
  group('DeliveryRules.validateTransition', () {
    test('allows assigned → en_route_store', () {
      final result = DeliveryRules.validateTransition(
        from: DeliveryLifecycleStatus.assigned,
        to: DeliveryLifecycleStatus.enRouteStore,
      );
      expect(result.isSuccess, isTrue);
    });

    test('blocks assigned → delivered', () {
      final result = DeliveryRules.validateTransition(
        from: DeliveryLifecycleStatus.assigned,
        to: DeliveryLifecycleStatus.delivered,
      );
      expect(result.isFailure, isTrue);
    });
  });

  group('DeliveryRules.validatePickupScan', () {
    test('succeeds at store with matching token', () {
      final result = DeliveryRules.validatePickupScan(
        status: DeliveryLifecycleStatus.atStore,
        scannedToken: 'token-1',
        expectedHandoffToken: 'token-1',
      );
      expect(result.isSuccess, isTrue);
    });

    test('fails on token mismatch', () {
      final result = DeliveryRules.validatePickupScan(
        status: DeliveryLifecycleStatus.atStore,
        scannedToken: 'wrong',
        expectedHandoffToken: 'token-1',
      );
      expect(result.isFailure, isTrue);
    });
  });

  group('DeliveryRules.validatePod', () {
    test('requires photo when arrived', () {
      final result = DeliveryRules.validatePod(
        status: DeliveryLifecycleStatus.arrived,
        hasPhoto: false,
        otp: '1234',
        otpRequired: true,
      );
      expect(result.isFailure, isTrue);
    });

    test('succeeds with photo and otp', () {
      final result = DeliveryRules.validatePod(
        status: DeliveryLifecycleStatus.arrived,
        hasPhoto: true,
        otp: '123456',
        otpRequired: true,
      );
      expect(result.isSuccess, isTrue);
    });
  });

  group('DeliveryRules.validateFailureReason', () {
    test('accepts known reason codes', () {
      expect(
        DeliveryRules.validateFailureReason('customer_unavailable').isSuccess,
        isTrue,
      );
    });

    test('rejects unknown reason', () {
      expect(
        DeliveryRules.validateFailureReason('not_a_reason').isFailure,
        isTrue,
      );
    });
  });
}
