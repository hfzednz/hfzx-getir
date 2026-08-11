import 'package:nexora_core/nexora_core.dart';
import '../../domain/entities/pack_task.dart';
import '../../domain/repositories/packing_repository.dart';
import '../datasources/packing_remote_datasource.dart';

class PackingRepositoryImpl implements PackingRepository {
  PackingRepositoryImpl(this._remote);
  final PackingRemoteDataSource _remote;

  @override
  Future<Result<List<PackTask>>> listQueue() => _remote.listQueue();
  @override
  Future<Result<PackTask>> getTask(String taskId) => _remote.getTask(taskId);
  @override
  Future<Result<PackTask>> claim(String taskId, {required String idempotencyKey}) =>
      _remote.claim(taskId, idempotencyKey: idempotencyKey);
  @override
  Future<Result<PackTask>> weigh(String taskId, {required double actualGrams, required String idempotencyKey}) =>
      _remote.weigh(taskId, actualGrams: actualGrams, idempotencyKey: idempotencyKey);
  @override
  Future<Result<PackTask>> printLabel(String taskId, {required String idempotencyKey}) =>
      _remote.printLabel(taskId, idempotencyKey: idempotencyKey);
  @override
  Future<Result<PackTask>> seal(String taskId, {required String idempotencyKey}) =>
      _remote.seal(taskId, idempotencyKey: idempotencyKey);
  @override
  Future<Result<PackTask>> complete(String taskId, {required String idempotencyKey}) =>
      _remote.complete(taskId, idempotencyKey: idempotencyKey);
}
