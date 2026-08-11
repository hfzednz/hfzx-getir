import 'package:dio/dio.dart';

import '../../config/nexora_headers.dart';

/// Attaches `Accept-Language` to every request (CONSTITUTION §30).
class AcceptLanguageInterceptor extends Interceptor {
  AcceptLanguageInterceptor({required this.languageProvider});

  final String Function() languageProvider;

  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) {
    options.headers.putIfAbsent(
      NexoraHeaders.acceptLanguage,
      languageProvider,
    );
    handler.next(options);
  }
}
