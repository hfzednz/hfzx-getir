import 'package:nexora_core/nexora_core.dart';
import '../entities/ops_task.dart';

abstract class TasksRepository {
  Future<Result<List<OpsTask>>> list();
  Future<Result<OpsTask>> complete(String id, {required String idempotencyKey});
}
