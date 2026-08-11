import 'package:equatable/equatable.dart';

import 'error_envelope.dart';
import 'nexora_error_code.dart';

/// Base domain exception for all NEXORA client failures.
sealed class NexoraException extends Equatable implements Exception {
  const NexoraException({
    required this.code,
    required this.message,
    this.details,
    this.traceId,
    this.supportCode,
    this.cause,
  });

  factory NexoraException.fromEnvelope(NexoraErrorEnvelope envelope) {
    return switch (envelope.code) {
      NexoraErrorCode.validationFailed ||
      NexoraErrorCode.invalidRequest =>
        NexoraValidationException(
          code: envelope.code,
          message: envelope.message,
          details: envelope.details,
          traceId: envelope.traceId,
          supportCode: envelope.supportCode,
        ),
      NexoraErrorCode.authRequired ||
      NexoraErrorCode.authInvalid ||
      NexoraErrorCode.authExpired ||
      NexoraErrorCode.refreshTokenReused =>
        NexoraAuthException(
          code: envelope.code,
          message: envelope.message,
          details: envelope.details,
          traceId: envelope.traceId,
          supportCode: envelope.supportCode,
        ),
      NexoraErrorCode.authForbidden => NexoraForbiddenException(
          code: envelope.code,
          message: envelope.message,
          details: envelope.details,
          traceId: envelope.traceId,
          supportCode: envelope.supportCode,
        ),
      NexoraErrorCode.notFound => NexoraNotFoundException(
          code: envelope.code,
          message: envelope.message,
          details: envelope.details,
          traceId: envelope.traceId,
          supportCode: envelope.supportCode,
        ),
      NexoraErrorCode.conflict => NexoraConflictException(
          code: envelope.code,
          message: envelope.message,
          details: envelope.details,
          traceId: envelope.traceId,
          supportCode: envelope.supportCode,
        ),
      NexoraErrorCode.idempotencyReplay => NexoraIdempotencyReplayException(
          code: envelope.code,
          message: envelope.message,
          details: envelope.details,
          traceId: envelope.traceId,
          supportCode: envelope.supportCode,
        ),
      NexoraErrorCode.rateLimited => NexoraRateLimitException(
          code: envelope.code,
          message: envelope.message,
          details: envelope.details,
          traceId: envelope.traceId,
          supportCode: envelope.supportCode,
        ),
      NexoraErrorCode.serviceUnavailable => NexoraServiceUnavailableException(
          code: envelope.code,
          message: envelope.message,
          details: envelope.details,
          traceId: envelope.traceId,
          supportCode: envelope.supportCode,
        ),
      _ => NexoraServerException(
          code: envelope.code,
          message: envelope.message,
          details: envelope.details,
          traceId: envelope.traceId,
          supportCode: envelope.supportCode,
        ),
    };
  }

  final NexoraErrorCode code;
  final String message;
  final dynamic details;
  final String? traceId;
  final String? supportCode;
  final Object? cause;

  @override
  List<Object?> get props =>
      [code, message, details, traceId, supportCode, cause];

  @override
  String toString() {
    final support = supportCode != null ? ' [$supportCode]' : '';
    return 'NexoraException(${code.code}): $message$support';
  }
}

final class NexoraValidationException extends NexoraException {
  const NexoraValidationException({
    required super.code,
    required super.message,
    super.details,
    super.traceId,
    super.supportCode,
    super.cause,
  });
}

final class NexoraAuthException extends NexoraException {
  const NexoraAuthException({
    required super.code,
    required super.message,
    super.details,
    super.traceId,
    super.supportCode,
    super.cause,
  });

  bool get requiresReLogin =>
      code == NexoraErrorCode.refreshTokenReused ||
      code == NexoraErrorCode.authInvalid;
}

final class NexoraForbiddenException extends NexoraException {
  const NexoraForbiddenException({
    required super.code,
    required super.message,
    super.details,
    super.traceId,
    super.supportCode,
    super.cause,
  });
}

final class NexoraNotFoundException extends NexoraException {
  const NexoraNotFoundException({
    required super.code,
    required super.message,
    super.details,
    super.traceId,
    super.supportCode,
    super.cause,
  });
}

final class NexoraConflictException extends NexoraException {
  const NexoraConflictException({
    required super.code,
    required super.message,
    super.details,
    super.traceId,
    super.supportCode,
    super.cause,
  });
}

final class NexoraIdempotencyReplayException extends NexoraException {
  const NexoraIdempotencyReplayException({
    required super.code,
    required super.message,
    super.details,
    super.traceId,
    super.supportCode,
    super.cause,
  });
}

final class NexoraRateLimitException extends NexoraException {
  const NexoraRateLimitException({
    required super.code,
    required super.message,
    super.details,
    super.traceId,
    super.supportCode,
    super.cause,
    this.retryAfter,
  });

  final Duration? retryAfter;
}

final class NexoraServiceUnavailableException extends NexoraException {
  const NexoraServiceUnavailableException({
    required super.code,
    required super.message,
    super.details,
    super.traceId,
    super.supportCode,
    super.cause,
  });
}

final class NexoraServerException extends NexoraException {
  const NexoraServerException({
    required super.code,
    required super.message,
    super.details,
    super.traceId,
    super.supportCode,
    super.cause,
  });
}

final class NexoraNetworkException extends NexoraException {
  const NexoraNetworkException({
    super.code = NexoraErrorCode.networkError,
    required super.message,
    super.details,
    super.traceId,
    super.supportCode,
    super.cause,
  });
}

final class NexoraTimeoutException extends NexoraException {
  const NexoraTimeoutException({
    super.code = NexoraErrorCode.timeout,
    required super.message,
    super.details,
    super.traceId,
    super.supportCode,
    super.cause,
  });
}

final class NexoraCancelledException extends NexoraException {
  const NexoraCancelledException({
    super.code = NexoraErrorCode.cancelled,
    super.message = 'Request was cancelled',
    super.details,
    super.traceId,
    super.supportCode,
    super.cause,
  });
}
