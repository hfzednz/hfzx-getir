import 'dart:async';
import 'dart:math';

import '../connectivity/connectivity_service.dart';
import '../errors/nexora_error_code.dart';
import '../errors/nexora_exception.dart';
import '../errors/result.dart';
import '../network/api_client.dart';
import 'models/pending_mutation.dart';
import 'mutation_outbox.dart';
import 'sync_engine.dart';

typedef MutationExecutor = Future<Result<void>> Function(PendingMutation mutation);

/// Default sync engine: drains [MutationOutbox] with exponential backoff + jitter.
class DefaultSyncEngine implements SyncEngine {
  DefaultSyncEngine({
    required MutationOutbox outbox,
    required ApiClient apiClient,
    required ConnectivityService connectivity,
    MutationExecutor? executor,
    this.maxRetries = 8,
    this.baseDelay = const Duration(seconds: 2),
    this.maxDelay = const Duration(minutes: 5),
  })  : _outbox = outbox,
        _connectivity = connectivity,
        _executor = executor ?? _defaultExecutor(apiClient);

  final MutationOutbox _outbox;
  final ConnectivityService _connectivity;
  final MutationExecutor _executor;

  final int maxRetries;
  final Duration baseDelay;
  final Duration maxDelay;

  final _events = StreamController<SyncEngineEvent>.broadcast();
  bool _flushing = false;

  @override
  Stream<SyncEngineEvent> get events => _events.stream;

  @override
  Future<void> flush() async {
    if (_flushing) {
      return;
    }
    if (!await _connectivity.isOnline) {
      return;
    }

    _flushing = true;
    var processed = 0;

    try {
      final pending = await _outbox.peek();
      for (final mutation in pending) {
        if (!await _connectivity.isOnline) {
          break;
        }

        final result = await _executor(mutation);
        if (result.isSuccess) {
          await _outbox.markSucceeded(mutation.clientMutationId);
          _events.add(SyncMutationSucceeded(mutation));
          processed += 1;
          continue;
        }

        final error = result.errorOrNull!;
        final nextRetry = mutation.retryCount + 1;

        if (nextRetry >= maxRetries || !_isRetriable(error)) {
          await _outbox.markFailed(
            mutation.clientMutationId,
            error: error.message,
            nextAttemptAt: DateTime.now().toUtc().add(maxDelay),
            retryCount: nextRetry,
          );
        } else {
          final delay = _computeDelay(nextRetry, error);
          await _outbox.markFailed(
            mutation.clientMutationId,
            error: error.message,
            nextAttemptAt: DateTime.now().toUtc().add(delay),
            retryCount: nextRetry,
          );
        }

        _events.add(SyncMutationFailed(mutation, error));
      }

      final remaining = await _outbox.length();
      _events.add(
        SyncFlushCompleted(processed: processed, remaining: remaining),
      );
    } finally {
      _flushing = false;
    }
  }

  static MutationExecutor _defaultExecutor(ApiClient apiClient) {
    return (mutation) async {
      final Result<void> result = switch (mutation.method.toUpperCase()) {
        'POST' => await apiClient.post<void>(
            mutation.path,
            data: mutation.body,
            idempotencyKey: mutation.idempotencyKey,
          ),
        'PUT' => await apiClient.put<void>(
            mutation.path,
            data: mutation.body,
            idempotencyKey: mutation.idempotencyKey,
          ),
        'PATCH' => await apiClient.patch<void>(
            mutation.path,
            data: mutation.body,
            idempotencyKey: mutation.idempotencyKey,
          ),
        'DELETE' => await apiClient.delete<void>(
            mutation.path,
            data: mutation.body,
            idempotencyKey: mutation.idempotencyKey,
          ),
        _ => Failure<void>(
            NexoraValidationException(
              code: NexoraErrorCode.invalidRequest,
              message: 'Unsupported mutation method: ${mutation.method}',
            ),
          ),
      };
      return result;
    };
  }

  Duration _computeDelay(int retryCount, NexoraException error) {
    if (error is NexoraRateLimitException && error.retryAfter != null) {
      return error.retryAfter!;
    }

    final exponent = min(retryCount, 10);
    final baseMs = baseDelay.inMilliseconds * pow(2, exponent).toInt();
    final jitter = Random().nextInt(400);
    final capped = min(baseMs + jitter, maxDelay.inMilliseconds);
    return Duration(milliseconds: capped);
  }

  bool _isRetriable(NexoraException error) {
    return error is NexoraNetworkException ||
        error is NexoraTimeoutException ||
        error is NexoraServiceUnavailableException ||
        error is NexoraRateLimitException;
  }

  Future<void> dispose() async {
    await _events.close();
  }
}
