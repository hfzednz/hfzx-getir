import 'package:dio/dio.dart';

import 'error_envelope.dart';
import 'nexora_error_code.dart';
import 'nexora_exception.dart';

/// Maps [DioException] and HTTP responses to [NexoraException] domain errors.
abstract final class DioErrorMapper {
  static NexoraException map(DioException error) {
    switch (error.type) {
      case DioExceptionType.cancel:
        return NexoraCancelledException(cause: error);
      case DioExceptionType.connectionTimeout:
      case DioExceptionType.sendTimeout:
      case DioExceptionType.receiveTimeout:
      case DioExceptionType.transformTimeout:
        return NexoraTimeoutException(
          message: 'Request timed out',
          cause: error,
        );
      case DioExceptionType.connectionError:
      case DioExceptionType.badCertificate:
        return NexoraNetworkException(
          message: 'Network connection failed',
          cause: error,
        );
      case DioExceptionType.badResponse:
        return _mapResponse(error);
      case DioExceptionType.unknown:
        return NexoraNetworkException(
          message: error.message ?? 'Unexpected network error',
          cause: error,
        );
    }
  }

  static NexoraException _mapResponse(DioException error) {
    final response = error.response;
    if (response == null) {
      return NexoraNetworkException(
        message: error.message ?? 'Empty response',
        cause: error,
      );
    }

    final statusCode = response.statusCode ?? 0;
    final data = response.data;

    NexoraErrorEnvelope? envelope;
    if (data is Map<String, dynamic>) {
      envelope = NexoraErrorEnvelope.fromJson(data);
    }

    if (envelope != null && envelope.code != NexoraErrorCode.unknown) {
      final mapped = NexoraException.fromEnvelope(envelope);
      if (mapped is NexoraRateLimitException) {
        return NexoraRateLimitException(
          code: mapped.code,
          message: mapped.message,
          details: mapped.details,
          traceId: mapped.traceId,
          supportCode: mapped.supportCode,
          cause: error,
          retryAfter: _parseRetryAfter(response.headers.value('retry-after')),
        );
      }
      return _withCause(mapped, error);
    }

    return switch (statusCode) {
      400 => NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: envelope?.message ?? 'Validation failed',
          traceId: envelope?.traceId,
          supportCode: envelope?.supportCode,
          cause: error,
        ),
      401 => NexoraAuthException(
          code: NexoraErrorCode.authRequired,
          message: envelope?.message ?? 'Authentication required',
          traceId: envelope?.traceId,
          supportCode: envelope?.supportCode,
          cause: error,
        ),
      403 => NexoraForbiddenException(
          code: NexoraErrorCode.authForbidden,
          message: envelope?.message ?? 'Access denied',
          traceId: envelope?.traceId,
          supportCode: envelope?.supportCode,
          cause: error,
        ),
      404 => NexoraNotFoundException(
          code: NexoraErrorCode.notFound,
          message: envelope?.message ?? 'Resource not found',
          traceId: envelope?.traceId,
          supportCode: envelope?.supportCode,
          cause: error,
        ),
      409 => NexoraConflictException(
          code: NexoraErrorCode.conflict,
          message: envelope?.message ?? 'Conflict',
          traceId: envelope?.traceId,
          supportCode: envelope?.supportCode,
          cause: error,
        ),
      429 => NexoraRateLimitException(
          code: NexoraErrorCode.rateLimited,
          message: envelope?.message ?? 'Too many requests',
          traceId: envelope?.traceId,
          supportCode: envelope?.supportCode,
          cause: error,
          retryAfter: _parseRetryAfter(response.headers.value('retry-after')),
        ),
      >= 500 && < 600 => NexoraServiceUnavailableException(
          code: NexoraErrorCode.serviceUnavailable,
          message: envelope?.message ?? 'Service unavailable',
          traceId: envelope?.traceId,
          supportCode: envelope?.supportCode,
          cause: error,
        ),
      _ => NexoraServerException(
          code: NexoraErrorCode.internalError,
          message: envelope?.message ?? 'Unexpected server error',
          traceId: envelope?.traceId,
          supportCode: envelope?.supportCode,
          cause: error,
        ),
    };
  }

  static NexoraException _withCause(NexoraException exception, DioException error) {
    return switch (exception) {
      NexoraValidationException() => NexoraValidationException(
          code: exception.code,
          message: exception.message,
          details: exception.details,
          traceId: exception.traceId,
          supportCode: exception.supportCode,
          cause: error,
        ),
      NexoraAuthException() => NexoraAuthException(
          code: exception.code,
          message: exception.message,
          details: exception.details,
          traceId: exception.traceId,
          supportCode: exception.supportCode,
          cause: error,
        ),
      NexoraForbiddenException() => NexoraForbiddenException(
          code: exception.code,
          message: exception.message,
          details: exception.details,
          traceId: exception.traceId,
          supportCode: exception.supportCode,
          cause: error,
        ),
      NexoraNotFoundException() => NexoraNotFoundException(
          code: exception.code,
          message: exception.message,
          details: exception.details,
          traceId: exception.traceId,
          supportCode: exception.supportCode,
          cause: error,
        ),
      NexoraConflictException() => NexoraConflictException(
          code: exception.code,
          message: exception.message,
          details: exception.details,
          traceId: exception.traceId,
          supportCode: exception.supportCode,
          cause: error,
        ),
      NexoraIdempotencyReplayException() => NexoraIdempotencyReplayException(
          code: exception.code,
          message: exception.message,
          details: exception.details,
          traceId: exception.traceId,
          supportCode: exception.supportCode,
          cause: error,
        ),
      NexoraRateLimitException() => NexoraRateLimitException(
          code: exception.code,
          message: exception.message,
          details: exception.details,
          traceId: exception.traceId,
          supportCode: exception.supportCode,
          cause: error,
          retryAfter: exception.retryAfter,
        ),
      NexoraServiceUnavailableException() => NexoraServiceUnavailableException(
          code: exception.code,
          message: exception.message,
          details: exception.details,
          traceId: exception.traceId,
          supportCode: exception.supportCode,
          cause: error,
        ),
      _ => NexoraServerException(
          code: exception.code,
          message: exception.message,
          details: exception.details,
          traceId: exception.traceId,
          supportCode: exception.supportCode,
          cause: error,
        ),
    };
  }

  static Duration? _parseRetryAfter(String? header) {
    if (header == null || header.isEmpty) {
      return null;
    }
    final seconds = int.tryParse(header);
    if (seconds != null) {
      return Duration(seconds: seconds);
    }
    return null;
  }
}
