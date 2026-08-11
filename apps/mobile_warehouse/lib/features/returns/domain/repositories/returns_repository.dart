import 'package:nexora_core/nexora_core.dart';
import '../entities/return_task.dart';

abstract class ReturnsRepository {
  Future<Result<List<ReturnTask>>> list({String? type});
  Future<Result<ReturnTask>> advance(String id, {required String action, required String idempotencyKey});
}
