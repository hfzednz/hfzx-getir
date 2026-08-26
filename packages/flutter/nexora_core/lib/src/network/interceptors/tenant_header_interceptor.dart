import 'package:dio/dio.dart';

import '../../config/nexora_headers.dart';

/// Injects `X-Tenant-Id` required by NEXORA BFFs.
class TenantHeaderInterceptor extends Interceptor {
  TenantHeaderInterceptor({required this.tenantIdProvider});

  final String Function() tenantIdProvider;

  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) {
    final tenantId = tenantIdProvider();
    if (tenantId.isNotEmpty) {
      options.headers[NexoraHeaders.tenantId] = tenantId;
    }
    handler.next(options);
  }
}
