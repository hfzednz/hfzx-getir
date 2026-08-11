import 'package:dio/dio.dart';
import 'package:uuid/uuid.dart';

import '../../config/nexora_headers.dart';

/// Adds `Idempotency-Key` to mutating requests (CONSTITUTION §30, ADR-007).
class IdempotencyInterceptor extends Interceptor {
  IdempotencyInterceptor({Uuid? uuid}) : _uuid = uuid ?? const Uuid();

  final Uuid _uuid;

  static const _mutatingMethods = {'POST', 'PUT', 'PATCH', 'DELETE'};

  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) {
    if (_mutatingMethods.contains(options.method.toUpperCase())) {
      options.headers.putIfAbsent(
        NexoraHeaders.idempotencyKey,
        () => options.extra['idempotency_key'] as String? ?? _uuid.v4(),
      );
    }
    handler.next(options);
  }
}
