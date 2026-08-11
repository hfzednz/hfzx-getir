import 'package:nexora_core/nexora_core.dart';

import '../entities/pick_task.dart';

abstract class PickingRepository {
  Future<Result<List<PickTask>>> listQueue();
  Future<Result<PickTask>> getTask(String taskId);
  Future<Result<PickTask>> claim(String taskId, {required String idempotencyKey});
  Future<Result<PickTask>> start(String taskId, {required String idempotencyKey});
  Future<Result<PickTask>> scanLine({
    required String taskId,
    required String lineId,
    required String barcode,
    required int qty,
    required String idempotencyKey,
  });
  Future<Result<PickTask>> shortPick({
    required String taskId,
    required String lineId,
    required int missingQty,
    required String idempotencyKey,
  });
  Future<Result<PickTask>> complete(String taskId, {required String idempotencyKey});
  Future<Result<PickTask>> stage(String taskId, {required String idempotencyKey});
}
