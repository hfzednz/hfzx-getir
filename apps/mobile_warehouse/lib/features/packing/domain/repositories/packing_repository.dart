import 'package:nexora_core/nexora_core.dart';
import '../entities/pack_task.dart';

abstract class PackingRepository {
  Future<Result<List<PackTask>>> listQueue();
  Future<Result<PackTask>> getTask(String taskId);
  Future<Result<PackTask>> claim(String taskId, {required String idempotencyKey});
  Future<Result<PackTask>> weigh(String taskId, {required double actualGrams, required String idempotencyKey});
  Future<Result<PackTask>> printLabel(String taskId, {required String idempotencyKey});
  Future<Result<PackTask>> seal(String taskId, {required String idempotencyKey});
  Future<Result<PackTask>> complete(String taskId, {required String idempotencyKey});
}
