import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nexora_core/nexora_core.dart';

void main() {
  group('DioErrorMapper', () {
    test('maps validation envelope to NexoraValidationException', () {
      final error = DioException(
        requestOptions: RequestOptions(path: '/v1/cart'),
        response: Response(
          requestOptions: RequestOptions(path: '/v1/cart'),
          statusCode: 400,
          data: {
            'error': {
              'code': 'VALIDATION_FAILED',
              'message': 'Invalid quantity',
              'trace_id': 'trace-123',
              'support_code': 'NX-4001',
              'details': [
                {'field': 'quantity', 'message': 'must be positive'},
              ],
            },
          },
        ),
        type: DioExceptionType.badResponse,
      );

      final mapped = DioErrorMapper.map(error);

      expect(mapped, isA<NexoraValidationException>());
      expect(mapped.code, NexoraErrorCode.validationFailed);
      expect(mapped.message, 'Invalid quantity');
      expect(mapped.traceId, 'trace-123');
      expect(mapped.supportCode, 'NX-4001');
    });

    test('maps refresh token reuse to NexoraAuthException', () {
      final error = DioException(
        requestOptions: RequestOptions(path: '/v1/auth/refresh'),
        response: Response(
          requestOptions: RequestOptions(path: '/v1/auth/refresh'),
          statusCode: 401,
          data: {
            'error': {
              'code': 'REFRESH_TOKEN_REUSED',
              'message': 'Session revoked',
            },
          },
        ),
        type: DioExceptionType.badResponse,
      );

      final mapped = DioErrorMapper.map(error);

      expect(mapped, isA<NexoraAuthException>());
      expect(mapped.code, NexoraErrorCode.refreshTokenReused);
      expect((mapped as NexoraAuthException).requiresReLogin, isTrue);
    });

    test('maps rate limit with Retry-After header', () {
      final error = DioException(
        requestOptions: RequestOptions(path: '/v1/search'),
        response: Response(
          requestOptions: RequestOptions(path: '/v1/search'),
          statusCode: 429,
          headers: Headers.fromMap({
            'retry-after': ['30'],
          }),
          data: {
            'error': {
              'code': 'RATE_LIMITED',
              'message': 'Slow down',
            },
          },
        ),
        type: DioExceptionType.badResponse,
      );

      final mapped = DioErrorMapper.map(error);

      expect(mapped, isA<NexoraRateLimitException>());
      expect((mapped as NexoraRateLimitException).retryAfter, const Duration(seconds: 30));
    });

    test('maps connection timeout to NexoraTimeoutException', () {
      final error = DioException(
        requestOptions: RequestOptions(path: '/v1/home'),
        type: DioExceptionType.connectionTimeout,
        message: 'timeout',
      );

      final mapped = DioErrorMapper.map(error);

      expect(mapped, isA<NexoraTimeoutException>());
      expect(mapped.code, NexoraErrorCode.timeout);
    });

    test('maps cancel to NexoraCancelledException', () {
      final error = DioException(
        requestOptions: RequestOptions(path: '/v1/home'),
        type: DioExceptionType.cancel,
      );

      final mapped = DioErrorMapper.map(error);

      expect(mapped, isA<NexoraCancelledException>());
    });
  });

  group('NexoraErrorEnvelope', () {
    test('fromJson parses constitution envelope', () {
      final envelope = NexoraErrorEnvelope.fromJson({
        'error': {
          'code': 'INV_STOCKOUT',
          'message': 'Some items are unavailable',
          'trace_id': 'abc',
          'support_code': 'NX-99',
        },
      });

      expect(envelope.code, NexoraErrorCode.unknown);
      expect(envelope.message, 'Some items are unavailable');
      expect(envelope.traceId, 'abc');
      expect(envelope.supportCode, 'NX-99');
    });
  });

  group('Result', () {
    test('fold handles success and failure', () {
      const success = Success<int>(42);
      const failure = Failure<int>(
        NexoraNetworkException(message: 'offline'),
      );

      expect(
        success.fold(onSuccess: (v) => v * 2, onFailure: (_) => -1),
        84,
      );
      expect(
        failure.fold(onSuccess: (_) => 1, onFailure: (e) => e.code.code.length),
        NexoraErrorCode.networkError.code.length,
      );
    });
  });
}
