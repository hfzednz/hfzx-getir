import 'package:nexora_core/nexora_core.dart';

import '../utils/idempotency.dart';

enum PaymentAttemptStatus {
  idle,
  inProgress,
  succeeded,
  failed,
  cancelled,
}

/// Tracks in-flight payment attempts for duplicate protection.
class PaymentAttemptState {
  const PaymentAttemptState({
    this.status = PaymentAttemptStatus.idle,
    this.idempotencyKey,
    this.paymentIntentId,
    this.lastAttemptAt,
    this.retryCount = 0,
  });

  final PaymentAttemptStatus status;
  final String? idempotencyKey;
  final String? paymentIntentId;
  final DateTime? lastAttemptAt;
  final int retryCount;

  PaymentAttemptState copyWith({
    PaymentAttemptStatus? status,
    String? idempotencyKey,
    String? paymentIntentId,
    DateTime? lastAttemptAt,
    int? retryCount,
  }) =>
      PaymentAttemptState(
        status: status ?? this.status,
        idempotencyKey: idempotencyKey ?? this.idempotencyKey,
        paymentIntentId: paymentIntentId ?? this.paymentIntentId,
        lastAttemptAt: lastAttemptAt ?? this.lastAttemptAt,
        retryCount: retryCount ?? this.retryCount,
      );
}

/// Payment idempotency and retry eligibility rules (CONSTITUTION §30).
abstract final class PaymentRules {
  static const maxRetryCount = 3;
  static const retryCooldown = Duration(seconds: 30);
  static const idempotencyKeyPattern = r'^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$';

  static final _uuidRegex = RegExp(idempotencyKeyPattern, caseSensitive: false);

  static Result<String> validateIdempotencyKey(String? key) {
    final value = key?.trim() ?? '';
    if (value.isEmpty) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Payment idempotency key is required',
          details: {'field': 'idempotency_key'},
        ),
      );
    }

    if (!_uuidRegex.hasMatch(value)) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Invalid payment idempotency key format',
          details: {'field': 'idempotency_key'},
        ),
      );
    }

    return Success(value);
  }

  static String ensureIdempotencyKey(String? existing) {
    final parsed = validateIdempotencyKey(existing);
    if (parsed.isSuccess) return parsed.valueOrNull!;
    return Idempotency.generate();
  }

  /// Returns [Success] with the key to use, or [Failure] when a duplicate in-flight
  /// attempt should be blocked.
  static Result<String> guardDuplicatePayment({
    required PaymentAttemptState state,
    required String proposedKey,
    DateTime? now,
  }) {
    final keyResult = validateIdempotencyKey(proposedKey);
    if (keyResult.isFailure) return keyResult;

    final key = keyResult.valueOrNull!;
    final clock = now ?? DateTime.now();

    if (state.status == PaymentAttemptStatus.inProgress &&
        state.idempotencyKey != null &&
        state.idempotencyKey != key) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.conflict,
          message: 'A payment is already in progress',
          details: {
            'existing_idempotency_key': state.idempotencyKey,
            'proposed_idempotency_key': key,
          },
        ),
      );
    }

    if (state.status == PaymentAttemptStatus.succeeded &&
        state.idempotencyKey == key) {
      return Failure(
        NexoraIdempotencyReplayException(
          code: NexoraErrorCode.idempotencyReplay,
          message: 'Payment already completed',
          details: {'idempotency_key': key, 'payment_intent_id': state.paymentIntentId},
        ),
      );
    }

    if (state.status == PaymentAttemptStatus.inProgress &&
        state.lastAttemptAt != null &&
        clock.difference(state.lastAttemptAt!) < retryCooldown) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Please wait before retrying payment',
          details: {
            'retry_after_seconds': retryCooldown.inSeconds,
            'idempotency_key': key,
          },
        ),
      );
    }

    return Success(key);
  }

  static Result<void> validateRetryEligibility({
    required PaymentAttemptState state,
    required NexoraException? lastError,
    DateTime? now,
  }) {
    if (state.status == PaymentAttemptStatus.succeeded) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Payment already succeeded',
          details: {'payment_intent_id': state.paymentIntentId},
        ),
      );
    }

    if (state.retryCount >= maxRetryCount) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Maximum payment retry attempts reached',
          details: {'retry_count': state.retryCount, 'max_retries': maxRetryCount},
        ),
      );
    }

    final clock = now ?? DateTime.now();
    if (state.lastAttemptAt != null &&
        clock.difference(state.lastAttemptAt!) < retryCooldown) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Please wait before retrying payment',
          details: {'retry_after_seconds': retryCooldown.inSeconds},
        ),
      );
    }

    if (lastError != null && !_isRetriable(lastError)) {
      return Failure(lastError);
    }

    return const Success(null);
  }

  static bool _isRetriable(NexoraException error) {
    return switch (error) {
      NexoraNetworkException() ||
      NexoraTimeoutException() ||
      NexoraServiceUnavailableException() ||
      NexoraRateLimitException() =>
        true,
      NexoraValidationException() ||
      NexoraAuthException() ||
      NexoraForbiddenException() ||
      NexoraNotFoundException() ||
      NexoraConflictException() ||
      NexoraIdempotencyReplayException() =>
        false,
      _ => error.code == NexoraErrorCode.internalError,
    };
  }
}
