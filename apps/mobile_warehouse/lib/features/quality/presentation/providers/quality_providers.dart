import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_core/nexora_core.dart';
import 'package:uuid/uuid.dart';
import '../../../../di/providers.dart';
import '../../data/datasources/quality_remote_datasource.dart';
import '../../data/repositories/quality_repository_impl.dart';
import '../../domain/entities/qc_inspection.dart';
import '../../domain/repositories/quality_repository.dart';

final qualityRemoteDataSourceProvider = Provider((ref) => QualityRemoteDataSource(ref.watch(apiClientProvider)));
final qualityRepositoryProvider = Provider<QualityRepository>((ref) => QualityRepositoryImpl(ref.watch(qualityRemoteDataSourceProvider)));
final qualityQueueProvider = FutureProvider.autoDispose<List<QcInspection>>((ref) async {
  final r = await ref.watch(qualityRepositoryProvider).listQueue();
  return r.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});
final qualityActionsProvider = Provider((ref) => QualityActions(ref));

class QualityActions {
  QualityActions(this._ref);
  final Ref _ref;
  final _uuid = const Uuid();
  Future<Result<QcInspection>> decide(String id, {required bool pass}) async {
    final r = await _ref.read(qualityRepositoryProvider).decide(id, pass: pass, idempotencyKey: _uuid.v4());
    if (r.isSuccess) _ref.invalidate(qualityQueueProvider);
    return r;
  }
}
