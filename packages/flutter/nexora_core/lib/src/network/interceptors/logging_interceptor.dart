import 'package:dio/dio.dart';
import 'package:logging/logging.dart';

import '../../config/nexora_headers.dart';

/// Structured HTTP logging without PII (CONSTITUTION §17).
class LoggingInterceptor extends Interceptor {
  LoggingInterceptor({Logger? logger}) : _logger = logger ?? Logger('NexoraHttp');

  final Logger _logger;

  static final _piiHeaderKeys = {
    NexoraHeaders.authorization.toLowerCase(),
    'cookie',
    'set-cookie',
    'x-api-key',
  };

  static final _piiPatterns = [
    RegExp(r'Bearer\s+\S+', caseSensitive: false),
    RegExp(r'\+?\d{10,15}'),
    RegExp(r'[\w.+-]+@[\w-]+\.[\w.-]+'),
  ];

  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) {
    _logger.info(
      '→ ${options.method} ${_sanitizeUrl(options.uri)}',
    );
    handler.next(options);
  }

  @override
  void onResponse(Response<dynamic> response, ResponseInterceptorHandler handler) {
    _logger.info(
      '← ${response.statusCode} ${_sanitizeUrl(response.requestOptions.uri)}',
    );
    handler.next(response);
  }

  @override
  void onError(DioException err, ErrorInterceptorHandler handler) {
    _logger.warning(
      '✕ ${err.response?.statusCode ?? 'ERR'} '
      '${_sanitizeUrl(err.requestOptions.uri)} '
      '${_sanitizeMessage(err.message)}',
    );
    handler.next(err);
  }

  String _sanitizeUrl(Uri uri) {
    final sanitizedQuery = uri.queryParameters.entries
        .map((entry) {
          final key = entry.key.toLowerCase();
          if (key.contains('token') ||
              key.contains('ticket') ||
              key.contains('phone') ||
              key.contains('email')) {
            return '${entry.key}=[REDACTED]';
          }
          return '${entry.key}=${entry.value}';
        })
        .join('&');

    if (sanitizedQuery.isEmpty) {
      return uri.path;
    }
    return '${uri.path}?$sanitizedQuery';
  }

  String _sanitizeMessage(String? message) {
    if (message == null) {
      return '';
    }
    var sanitized = message;
    for (final pattern in _piiPatterns) {
      sanitized = sanitized.replaceAll(pattern, '[REDACTED]');
    }
    return sanitized;
  }

  static Map<String, dynamic> sanitizeHeaders(Map<String, dynamic> headers) {
    return headers.map((key, value) {
      if (_piiHeaderKeys.contains(key.toLowerCase())) {
        return MapEntry(key, '[REDACTED]');
      }
      return MapEntry(key, value);
    });
  }
}
