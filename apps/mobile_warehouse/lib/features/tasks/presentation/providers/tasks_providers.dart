import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_core/nexora_core.dart';
import 'package:uuid/uuid.dart';
import '../../../../di/providers.dart';
import '../../data/datasources/tasks_remote_datasource.dart';
import '../../data/repositories/tasks_repository_impl.dart';
import '../../domain/entities/ops_task.dart';
import '../../domain/repositories/tasks_repository.dart';

final tasksRemoteDataSourceProvider = Provider((ref) => TasksRemoteDataSource(ref.watch(apiClientProvider)));
final tasksRepositoryProvider = Provider<TasksRepository>((ref) => TasksRepositoryImpl(ref.watch(tasksRemoteDataSourceProvider)));
final tasksListProvider = FutureProvider.autoDispose<List<OpsTask>>((ref) async {
  final r = await ref.watch(tasksRepositoryProvider).list();
  return r.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});
final tasksActionsProvider = Provider((ref) => TasksActions(ref));

class TasksActions {
  TasksActions(this._ref);
  final Ref _ref;
  final _uuid = const Uuid();
  Future<Result<OpsTask>> complete(String id) async {
    final r = await _ref.read(tasksRepositoryProvider).complete(id, idempotencyKey: _uuid.v4());
    if (r.isSuccess) _ref.invalidate(tasksListProvider);
    return r;
  }
}
