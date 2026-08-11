import 'package:dio/dio.dart';

import '../../config/nexora_headers.dart';

/// Injects `X-Nexora-City-Id` when a city context is available (CONSTITUTION §30).
class CityHeaderInterceptor extends Interceptor {
  CityHeaderInterceptor({required this.cityIdProvider});

  final String? Function() cityIdProvider;

  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) {
    final cityId = cityIdProvider();
    if (cityId != null && cityId.isNotEmpty) {
      options.headers[NexoraHeaders.cityId] = cityId;
    }
    handler.next(options);
  }
}
