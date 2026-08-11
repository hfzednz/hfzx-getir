import 'package:flutter_test/flutter_test.dart';
import 'package:nexora_core/nexora_core.dart';
import 'package:nexora_customer/shared/business_rules/payment_rules.dart';

void main() {
  group('PaymentRules.validateIdempotencyKey', () {
    test('rejects empty key', () {
      final result = PaymentRules.validateIdempotencyKey('');
      expect(result.isFailure, isTrue);
    });

    test('accepts uuid key', () {
      final result = PaymentRules.validateIdempotencyKey(
        '550e8400-e29b-41d4-a716-446655440000',
      );
      expect(result.isSuccess, isTrue);
    });
  });

  group('PaymentRules.validateRetryEligibility', () {
    test('blocks retry after succeeded payment', () {
      final result = PaymentRules.validateRetryEligibility(
        state: const PaymentAttemptState(
          status: PaymentAttemptStatus.succeeded,
          paymentIntentId: 'pi_1',
        ),
        lastError: null,
      );
      expect(result.isFailure, isTrue);
      expect(result.errorOrNull!.message, contains('already succeeded'));
    });

    test('blocks when max retry count reached', () {
      final result = PaymentRules.validateRetryEligibility(
        state: const PaymentAttemptState(
          status: PaymentAttemptStatus.failed,
          retryCount: PaymentRules.maxRetryCount,
        ),
        lastError: const NexoraNetworkException(message: 'offline'),
      );
      expect(result.isFailure, isTrue);
      expect(result.errorOrNull!.message, contains('Maximum payment retry'));
    });

    test('blocks when cooldown has not elapsed', () {
      final now = DateTime.utc(2026, 1, 1, 12, 0, 0);
      final result = PaymentRules.validateRetryEligibility(
        state: PaymentAttemptState(
          status: PaymentAttemptStatus.failed,
          retryCount: 1,
          lastAttemptAt: now.subtract(const Duration(seconds: 5)),
        ),
        lastError: const NexoraNetworkException(message: 'offline'),
        now: now,
      );
      expect(result.isFailure, isTrue);
      expect(result.errorOrNull!.message, contains('Please wait'));
    });

    test('allows retry for retriable network error after cooldown', () {
      final now = DateTime.utc(2026, 1, 1, 12, 0, 0);
      final result = PaymentRules.validateRetryEligibility(
        state: PaymentAttemptState(
          status: PaymentAttemptStatus.failed,
          retryCount: 1,
          lastAttemptAt: now.subtract(PaymentRules.retryCooldown),
        ),
        lastError: const NexoraNetworkException(message: 'offline'),
        now: now,
      );
      expect(result.isSuccess, isTrue);
    });

    test('rejects non-retriable validation error', () {
      final now = DateTime.utc(2026, 1, 1, 12, 0, 0);
      final lastError = NexoraValidationException(
        code: NexoraErrorCode.validationFailed,
        message: 'Card declined',
      );
      final result = PaymentRules.validateRetryEligibility(
        state: PaymentAttemptState(
          status: PaymentAttemptStatus.failed,
          retryCount: 0,
          lastAttemptAt: now.subtract(PaymentRules.retryCooldown),
        ),
        lastError: lastError,
        now: now,
      );
      expect(result.isFailure, isTrue);
      expect(result.errorOrNull, same(lastError));
    });
  });

  group('PaymentRules.guardDuplicatePayment', () {
    test('blocks different key while in progress', () {
      final result = PaymentRules.guardDuplicatePayment(
        state: const PaymentAttemptState(
          status: PaymentAttemptStatus.inProgress,
          idempotencyKey: '550e8400-e29b-41d4-a716-446655440000',
        ),
        proposedKey: '550e8400-e29b-41d4-a716-446655440001',
      );
      expect(result.isFailure, isTrue);
      expect(result.errorOrNull!.message, contains('already in progress'));
    });
  });
}
