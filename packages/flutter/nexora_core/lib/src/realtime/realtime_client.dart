import 'dart:async';
import 'dart:math';

import 'package:logging/logging.dart';
import 'package:web_socket_channel/web_socket_channel.dart';

/// WebSocket client for NEXORA realtime-gateway (CONSTITUTION §29, §30).
class RealtimeClient {
  RealtimeClient({
    required this.wsBaseUrl,
    required this.ticketProvider,
    this.backoffInitial = const Duration(seconds: 1),
    this.backoffMax = const Duration(seconds: 30),
    this.logger,
  });

  final String wsBaseUrl;
  final Future<String> Function() ticketProvider;
  final Duration backoffInitial;
  final Duration backoffMax;
  final Logger? logger;

  WebSocketChannel? _channel;
  StreamSubscription<dynamic>? _subscription;
  Timer? _reconnectTimer;
  int _attempt = 0;
  bool _disposed = false;
  bool _manualDisconnect = false;

  final _controller = StreamController<RealtimeEvent>.broadcast();
  RealtimeConnectionState _state = RealtimeConnectionState.disconnected;

  Stream<RealtimeEvent> get events => _controller.stream;
  RealtimeConnectionState get state => _state;

  Future<void> connect() async {
    _manualDisconnect = false;
    await _openConnection();
  }

  Future<void> disconnect() async {
    _manualDisconnect = true;
    _reconnectTimer?.cancel();
    await _subscription?.cancel();
    await _channel?.sink.close();
    _channel = null;
    _setState(RealtimeConnectionState.disconnected);
  }

  Future<void> dispose() async {
    _disposed = true;
    await disconnect();
    await _controller.close();
  }

  Future<void> _openConnection() async {
    if (_disposed || _manualDisconnect) {
      return;
    }

    _setState(RealtimeConnectionState.connecting);

    try {
      final ticket = await ticketProvider();
      final uri = Uri.parse(wsBaseUrl).replace(
        queryParameters: {
          ...Uri.parse(wsBaseUrl).queryParameters,
          'ticket': ticket,
        },
      );

      _channel = WebSocketChannel.connect(uri);
      _subscription = _channel!.stream.listen(
        _onMessage,
        onError: _onError,
        onDone: _onDone,
        cancelOnError: true,
      );

      _attempt = 0;
      _setState(RealtimeConnectionState.connected);
      logger?.info('Realtime connected');
    } catch (error, stackTrace) {
      logger?.warning('Realtime connect failed', error, stackTrace);
      _scheduleReconnect();
    }
  }

  void _onMessage(dynamic message) {
    _controller.add(
      RealtimeEvent.raw(message is String ? message : message.toString()),
    );
  }

  void _onError(Object error, [StackTrace? stackTrace]) {
    logger?.warning('Realtime stream error', error, stackTrace);
    _scheduleReconnect();
  }

  void _onDone() {
    if (!_manualDisconnect && !_disposed) {
      _scheduleReconnect();
    }
  }

  void _scheduleReconnect() {
    if (_disposed || _manualDisconnect) {
      return;
    }

    _setState(RealtimeConnectionState.reconnecting);
    _reconnectTimer?.cancel();

    final delay = _backoffDelay();
    _attempt += 1;
    logger?.info('Realtime reconnect in ${delay.inMilliseconds}ms');

    _reconnectTimer = Timer(delay, () {
      unawaited(_openConnection());
    });
  }

  Duration _backoffDelay() {
    final exponent = min(_attempt, 10);
    final baseMs = backoffInitial.inMilliseconds * pow(2, exponent).toInt();
    final jitter = Random().nextInt(500);
    final capped = min(baseMs + jitter, backoffMax.inMilliseconds);
    return Duration(milliseconds: capped);
  }

  void _setState(RealtimeConnectionState next) {
    _state = next;
    _controller.add(RealtimeEvent.stateChanged(next));
  }
}

enum RealtimeConnectionState {
  disconnected,
  connecting,
  connected,
  reconnecting,
}

sealed class RealtimeEvent {
  const RealtimeEvent();

  factory RealtimeEvent.raw(String payload) = RealtimeMessageEvent;
  factory RealtimeEvent.stateChanged(RealtimeConnectionState state) =
      RealtimeStateEvent;
}

final class RealtimeMessageEvent extends RealtimeEvent {
  const RealtimeMessageEvent(this.payload);

  final String payload;
}

final class RealtimeStateEvent extends RealtimeEvent {
  const RealtimeStateEvent(this.state);

  final RealtimeConnectionState state;
}
