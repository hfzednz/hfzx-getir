import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_core/nexora_core.dart';
import 'package:uuid/uuid.dart';

import '../../../../di/providers.dart';
import '../../../../shared/business_rules/packing_rules.dart';
import '../../data/datasources/packing_remote_datasource.dart';
import '../../data/repositories/packing_repository_impl.dart';
import '../../domain/entities/pack_task.dart';
import '../../domain/repositories/packing_repository.dart';

final packingRemoteDataSourceProvider = Provider<PackingRemoteDataSource>((ref) {
  return PackingRemoteDataSource(ref.watch(apiClientProvider));
});

final packingRepositoryProvider = Provider<PackingRepository>((ref) {
  return PackingRepositoryImpl(ref.watch(packingRemoteDataSourceProvider));
});

final packingQueueProvider = FutureProvider.autoDispose<List<PackTask>>((ref) async {
  final result = await ref.watch(packingRepositoryProvider).listQueue();
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});

final packTaskProvider = FutureProvider.autoDispose.family<PackTask, String>((ref, id) async {
  final result = await ref.watch(packingRepositoryProvider).getTask(id);
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});

final packingActionsProvider = Provider((ref) => PackingActions(ref));

class PackingActions {
  PackingActions(this._ref);
  final Ref _ref;
  final _uuid = const Uuid();

  PackingRepository get _repo => _ref.read(packingRepositoryProvider);

  void _inv(String id) {
    _ref.invalidate(packingQueueProvider);
    _ref.invalidate(packTaskProvider(id));
  }

  Future<Result<PackTask>> claim(String id) async {
    final r = await _repo.claim(id, idempotencyKey: _uuid.v4());
    if (r.isSuccess) _inv(id);
    return r;
  }

  Future<Result<PackTask>> weigh(PackTask task, double actualGrams) async {
    final v = PackingRules.validateWeight(
      actualGrams: actualGrams,
      expectedGrams: task.expectedWeightGrams,
    );
    if (v.isFailure) return Failure(v.errorOrNull!);
    final r = await _repo.weigh(task.id, actualGrams: actualGrams, idempotencyKey: _uuid.v4());
    if (r.isSuccess) _inv(task.id);
    return r;
  }

  Future<Result<PackTask>> printLabel(String id) async {
    final r = await _repo.printLabel(id, idempotencyKey: _uuid.v4());
    if (r.isSuccess) _inv(id);
    return r;
  }

  Future<Result<PackTask>> seal(PackTask task) async {
    final v = PackingRules.validateSeal(
      status: task.status,
      sealed: true,
      labelPrinted: task.labelPrinted,
    );
    if (v.isFailure) return Failure(v.errorOrNull!);
    final r = await _repo.seal(task.id, idempotencyKey: _uuid.v4());
    if (r.isSuccess) _inv(task.id);
    return r;
  }

  Future<Result<PackTask>> complete(String id) async {
    final r = await _repo.complete(id, idempotencyKey: _uuid.v4());
    if (r.isSuccess) _inv(id);
    return r;
  }
}
