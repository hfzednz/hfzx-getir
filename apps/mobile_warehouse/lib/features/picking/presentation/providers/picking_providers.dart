import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_core/nexora_core.dart';
import 'package:uuid/uuid.dart';

import '../../../../di/providers.dart';
import '../../../../shared/business_rules/picking_rules.dart';
import '../../data/datasources/picking_remote_datasource.dart';
import '../../data/repositories/picking_repository_impl.dart';
import '../../domain/entities/pick_task.dart';
import '../../domain/repositories/picking_repository.dart';

final pickingRemoteDataSourceProvider = Provider<PickingRemoteDataSource>((ref) {
  return PickingRemoteDataSource(ref.watch(apiClientProvider));
});

final pickingRepositoryProvider = Provider<PickingRepository>((ref) {
  return PickingRepositoryImpl(ref.watch(pickingRemoteDataSourceProvider));
});

final pickingQueueProvider =
    FutureProvider.autoDispose<List<PickTask>>((ref) async {
  final result = await ref.watch(pickingRepositoryProvider).listQueue();
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});

final pickTaskProvider =
    FutureProvider.autoDispose.family<PickTask, String>((ref, taskId) async {
  final result = await ref.watch(pickingRepositoryProvider).getTask(taskId);
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});

final pickingActionsProvider = Provider((ref) => PickingActions(ref));

class PickingActions {
  PickingActions(this._ref);
  final Ref _ref;
  final _uuid = const Uuid();

  PickingRepository get _repo => _ref.read(pickingRepositoryProvider);

  void _invalidate(String taskId) {
    _ref.invalidate(pickingQueueProvider);
    _ref.invalidate(pickTaskProvider(taskId));
  }

  Future<Result<PickTask>> claim(String taskId) async {
    final result = await _repo.claim(taskId, idempotencyKey: _uuid.v4());
    if (result.isSuccess) _invalidate(taskId);
    return result;
  }

  Future<Result<PickTask>> start(String taskId) async {
    final result = await _repo.start(taskId, idempotencyKey: _uuid.v4());
    if (result.isSuccess) _invalidate(taskId);
    return result;
  }

  Future<Result<PickTask>> scanLine({
    required PickTask task,
    required PickLine line,
    required String barcode,
    int qty = 1,
  }) async {
    final validation = PickingRules.validateLineScan(
      status: task.status,
      line: line,
      scannedBarcode: barcode,
      qtyDelta: qty,
    );
    if (validation.isFailure) {
      return Failure(validation.errorOrNull!);
    }
    final result = await _repo.scanLine(
      taskId: task.id,
      lineId: line.id,
      barcode: barcode,
      qty: qty,
      idempotencyKey: _uuid.v4(),
    );
    if (result.isSuccess) _invalidate(task.id);
    return result;
  }

  Future<Result<PickTask>> shortPick({
    required PickTask task,
    required PickLine line,
    required int missingQty,
  }) async {
    final validation = PickingRules.validateShortPick(
      status: task.status,
      line: line,
      missingQty: missingQty,
    );
    if (validation.isFailure) {
      return Failure(validation.errorOrNull!);
    }
    final result = await _repo.shortPick(
      taskId: task.id,
      lineId: line.id,
      missingQty: missingQty,
      idempotencyKey: _uuid.v4(),
    );
    if (result.isSuccess) _invalidate(task.id);
    return result;
  }

  Future<Result<PickTask>> complete(String taskId) async {
    final result = await _repo.complete(taskId, idempotencyKey: _uuid.v4());
    if (result.isSuccess) _invalidate(taskId);
    return result;
  }

  Future<Result<PickTask>> stage(String taskId) async {
    final result = await _repo.stage(taskId, idempotencyKey: _uuid.v4());
    if (result.isSuccess) _invalidate(taskId);
    return result;
  }
}
