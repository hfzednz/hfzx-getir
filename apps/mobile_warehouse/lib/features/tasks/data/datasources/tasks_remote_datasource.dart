import 'package:nexora_core/nexora_core.dart';
import '../../domain/entities/ops_task.dart';

class TasksRemoteDataSource {
  const TasksRemoteDataSource(this._client);
  final ApiClient _client;
  Future<Result<List<OpsTask>>> list() {
    return _client.get<List<OpsTask>>(
      '/warehouse/tasks',
      parser: (json) {
        final list = json is List ? json : (json as Map)['items'] as List? ?? const [];
        return list.map((e) => OpsTask.fromJson(Map<String, dynamic>.from(e as Map))).toList();
      },
    );
  }
  Future<Result<OpsTask>> complete(String id, {required String idempotencyKey}) {
    return _client.post<OpsTask>(
      '/warehouse/tasks/$id/complete',
      idempotencyKey: idempotencyKey,
      parser: (json) => OpsTask.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }
}
