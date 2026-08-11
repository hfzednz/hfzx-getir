import 'dart:async';
import 'dart:convert';

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_core/nexora_core.dart';
import 'package:uuid/uuid.dart';

import '../../../../di/providers.dart';
import '../../../../shared/business_rules/handoff_rules.dart';
import '../../data/datasources/dispatch_remote_datasource.dart';
import '../../data/repositories/dispatch_repository_impl.dart';
import '../../domain/entities/handoff_task.dart';
import '../../domain/repositories/dispatch_repository.dart';

final dispatchRemoteDataSourceProvider = Provider<DispatchRemoteDataSource>((ref) {
  return DispatchRemoteDataSource(ref.watch(apiClientProvider));
});

final dispatchRepositoryProvider = Provider<DispatchRepository>((ref) {
  return DispatchRepositoryImpl(ref.watch(dispatchRemoteDataSourceProvider));
});

final dispatchQueueProvider = FutureProvider.autoDispose<List<HandoffTask>>((ref) async {
  final result = await ref.watch(dispatchRepositoryProvider).listQueue();
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});

final handoffProvider = FutureProvider.autoDispose.family<HandoffTask, String>((ref, id) async {
  final result = await ref.watch(dispatchRepositoryProvider).getHandoff(id);
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});

/// Invalidates dispatch queue on realtime `handoff.*` events.
final handoffRealtimeInvalidationProvider = Provider.autoDispose<void>((ref) {
  final client = ref.watch(realtimeClientProvider);
  unawaited(client.connect());
  final sub = client.events.listen((event) {
    if (event is! RealtimeMessageEvent) return;
    if (_isHandoffEvent(event.payload)) {
      ref.invalidate(dispatchQueueProvider);
    }
  });
  ref.onDispose(sub.cancel);
});

bool _isHandoffEvent(String payload) {
  try {
    final decoded = jsonDecode(payload);
    if (decoded is Map) {
      final type = decoded['type']?.toString() ??
          decoded['event']?.toString() ??
          decoded['event_type']?.toString() ??
          '';
      return type.startsWith('handoff.');
    }
  } catch (_) {}
  return payload.contains('handoff.');
}

final dispatchActionsProvider = Provider((ref) => DispatchActions(ref));

class DispatchActions {
  DispatchActions(this._ref);
  final Ref _ref;
  final _uuid = const Uuid();

  DispatchRepository get _repo => _ref.read(dispatchRepositoryProvider);

  void _inv(String id) {
    _ref.invalidate(dispatchQueueProvider);
    _ref.invalidate(handoffProvider(id));
  }

  Future<Result<HandoffTask>> markArrived(String id) async {
    final r = await _repo.markCourierArrived(id, idempotencyKey: _uuid.v4());
    if (r.isSuccess) _inv(id);
    return r;
  }

  Future<Result<HandoffTask>> scan({
    required HandoffTask task,
    required String scannedToken,
    String? scannedOrderId,
  }) async {
    final v = HandoffRules.validateHandoffScan(
      status: task.status,
      scannedToken: scannedToken,
      expectedToken: task.handoffToken,
      expectedOrderId: task.orderId,
      scannedOrderId: scannedOrderId,
    );
    if (v.isFailure) return Failure(v.errorOrNull!);
    final r = await _repo.scanHandoff(
      id: task.id,
      scannedToken: scannedToken,
      scannedOrderId: scannedOrderId,
      idempotencyKey: _uuid.v4(),
    );
    if (r.isSuccess) _inv(task.id);
    return r;
  }

  Future<Result<HandoffTask>> confirm(String id) async {
    final r = await _repo.confirm(id, idempotencyKey: _uuid.v4());
    if (r.isSuccess) _inv(id);
    return r;
  }

  Future<Result<HandoffTask>> fail(String id, {required String reasonCode, String? notes}) async {
    final v = HandoffRules.validateFailReason(reasonCode, notes: notes);
    if (v.isFailure) return Failure(v.errorOrNull!);
    final r = await _repo.fail(id, reasonCode: reasonCode, notes: notes, idempotencyKey: _uuid.v4());
    if (r.isSuccess) _inv(id);
    return r;
  }
}
