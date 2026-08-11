import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_core/nexora_core.dart';
import 'package:uuid/uuid.dart';
import '../../../../di/providers.dart';
import '../../data/datasources/returns_remote_datasource.dart';
import '../../data/repositories/returns_repository_impl.dart';
import '../../domain/entities/return_task.dart';
import '../../domain/repositories/returns_repository.dart';

final returnsRemoteDataSourceProvider = Provider((ref) => ReturnsRemoteDataSource(ref.watch(apiClientProvider)));
final returnsRepositoryProvider = Provider<ReturnsRepository>((ref) => ReturnsRepositoryImpl(ref.watch(returnsRemoteDataSourceProvider)));
final returnsListProvider = FutureProvider.autoDispose<List<ReturnTask>>((ref) async {
  final r = await ref.watch(returnsRepositoryProvider).list();
  return r.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});
final returnsActionsProvider = Provider((ref) => ReturnsActions(ref));

class ReturnsActions {
  ReturnsActions(this._ref);
  final Ref _ref;
  final _uuid = const Uuid();
  Future<Result<ReturnTask>> advance(String id, String action) async {
    final r = await _ref.read(returnsRepositoryProvider).advance(id, action: action, idempotencyKey: _uuid.v4());
    if (r.isSuccess) _ref.invalidate(returnsListProvider);
    return r;
  }
}
