import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/pick_task.dart';
import '../../domain/repositories/picking_repository.dart';
import '../datasources/picking_remote_datasource.dart';

class PickingRepositoryImpl implements PickingRepository {
  PickingRepositoryImpl(this._remote);
  final PickingRemoteDataSource _remote;

  @override
  Future<Result<List<PickTask>>> listQueue() => _remote.listQueue();

  @override
  Future<Result<PickTask>> getTask(String taskId) => _remote.getTask(taskId);

  @override
  Future<Result<PickTask>> claim(String taskId, {required String idempotencyKey}) =>
      _remote.claim(taskId, idempotencyKey: idempotencyKey);

  @override
  Future<Result<PickTask>> start(String taskId, {required String idempotencyKey}) =>
      _remote.start(taskId, idempotencyKey: idempotencyKey);

  @override
  Future<Result<PickTask>> scanLine({
    required String taskId,
    required String lineId,
    required String barcode,
    required int qty,
    required String idempotencyKey,
  }) =>
      _remote.scanLine(
        taskId: taskId,
        lineId: lineId,
        barcode: barcode,
        qty: qty,
        idempotencyKey: idempotencyKey,
      );

  @override
  Future<Result<PickTask>> shortPick({
    required String taskId,
    required String lineId,
    required int missingQty,
    required String idempotencyKey,
  }) =>
      _remote.shortPick(
        taskId: taskId,
        lineId: lineId,
        missingQty: missingQty,
        idempotencyKey: idempotencyKey,
      );

  @override
  Future<Result<PickTask>> complete(String taskId, {required String idempotencyKey}) =>
      _remote.complete(taskId, idempotencyKey: idempotencyKey);

  @override
  Future<Result<PickTask>> stage(String taskId, {required String idempotencyKey}) =>
      _remote.stage(taskId, idempotencyKey: idempotencyKey);
}
