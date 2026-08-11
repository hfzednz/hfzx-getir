import 'package:nexora_core/nexora_core.dart';
import '../../domain/entities/ops_task.dart';
import '../../domain/repositories/tasks_repository.dart';
import '../datasources/tasks_remote_datasource.dart';

class TasksRepositoryImpl implements TasksRepository {
  TasksRepositoryImpl(this._remote);
  final TasksRemoteDataSource _remote;
  @override
  Future<Result<List<OpsTask>>> list() => _remote.list();
  @override
  Future<Result<OpsTask>> complete(String id, {required String idempotencyKey}) =>
      _remote.complete(id, idempotencyKey: idempotencyKey);
}
