import 'package:logging/logging.dart';

/// Analytics gateway contract (CONSTITUTION §19).
abstract class AnalyticsGateway {
  Future<void> track({
    required String eventName,
    required String eventVersion,
    Map<String, Object?> props = const {},
    String? cityId,
    String? sessionId,
  });

  Future<void> flush();
}

/// No-op analytics sink for production builds that disable client telemetry.
class NoOpAnalyticsGateway implements AnalyticsGateway {
  const NoOpAnalyticsGateway();

  @override
  Future<void> flush() async {}

  @override
  Future<void> track({
    required String eventName,
    required String eventVersion,
    Map<String, Object?> props = const {},
    String? cityId,
    String? sessionId,
  }) async {}
}

/// Development analytics sink that logs sanitized events locally.
class LoggingAnalyticsGateway implements AnalyticsGateway {
  LoggingAnalyticsGateway({Logger? logger})
      : _logger = logger ?? Logger('Analytics');

  final Logger _logger;
  final List<Map<String, Object?>> _buffer = [];

  @override
  Future<void> track({
    required String eventName,
    required String eventVersion,
    Map<String, Object?> props = const {},
    String? cityId,
    String? sessionId,
  }) async {
    final event = <String, Object?>{
      'event_name': eventName,
      'event_version': eventVersion,
      'occurred_at': DateTime.now().toUtc().toIso8601String(),
      if (cityId != null) 'city_id': cityId,
      if (sessionId != null) 'session_id': sessionId,
      'props': _sanitizeProps(props),
    };
    _buffer.add(event);
    _logger.info('analytics: $eventName v$eventVersion');
  }

  @override
  Future<void> flush() async {
    if (_buffer.isEmpty) {
      return;
    }
    _logger.fine('analytics flush (${_buffer.length} events)');
    _buffer.clear();
  }

  Map<String, Object?> _sanitizeProps(Map<String, Object?> props) {
    const sensitiveKeys = {'email', 'phone', 'address', 'token', 'password'};
    return props.map((key, value) {
      if (sensitiveKeys.contains(key.toLowerCase())) {
        return MapEntry(key, '[REDACTED]');
      }
      return MapEntry(key, value);
    });
  }
}
