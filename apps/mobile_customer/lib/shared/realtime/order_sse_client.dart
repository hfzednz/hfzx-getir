import 'dart:async';
import 'dart:convert';

import 'package:dio/dio.dart';

import 'sse_parser.dart';

/// Authenticated HTTP SSE client for the realtime-gateway
/// `GET /v1/realtime/sse?topic=&ticket=` contract.
class OrderSseClient {
  OrderSseClient({Dio? dio}) : _dio = dio ?? Dio();

  final Dio _dio;

  Stream<SseFrame> connect({
    required String sseUrl,
    required String ticket,
    required String topic,
    Duration idleTimeout = const Duration(seconds: 45),
  }) {
    final controller = StreamController<SseFrame>();
    CancelToken? cancel;
    var closed = false;

    Future<void> run() async {
      var attempt = 0;
      while (!closed) {
        cancel = CancelToken();
        try {
          final response = await _dio.get<ResponseBody>(
            sseUrl,
            queryParameters: {'ticket': ticket, 'topic': topic},
            options: Options(
              responseType: ResponseType.stream,
              headers: const {
                'Accept': 'text/event-stream',
                'Cache-Control': 'no-cache',
              },
              receiveTimeout: idleTimeout,
              sendTimeout: const Duration(seconds: 15),
            ),
            cancelToken: cancel,
          );
          final stream = response.data?.stream;
          if (stream == null) {
            throw StateError('empty sse body');
          }
          attempt = 0;
          var pending = '';
          await for (final chunk in stream) {
            if (closed) return;
            pending += utf8.decode(chunk, allowMalformed: true);
            final parsed = parseSseBuffer(pending);
            pending = parsed.rest;
            for (final frame in parsed.frames) {
              if (!controller.isClosed) controller.add(frame);
            }
          }
        } catch (err, stack) {
          if (closed || cancel?.isCancelled == true) return;
          if (!controller.isClosed) controller.addError(err, stack);
        }
        if (closed) return;
        attempt++;
        final delay = Duration(milliseconds: (400 * (1 << (attempt.clamp(0, 5)))).clamp(400, 8000));
        await Future<void>.delayed(delay);
      }
    }

    controller.onListen = () {
      unawaited(run());
    };
    controller.onCancel = () {
      closed = true;
      cancel?.cancel();
    };
    return controller.stream;
  }
}
