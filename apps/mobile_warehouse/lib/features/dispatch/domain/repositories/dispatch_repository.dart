import 'package:nexora_core/nexora_core.dart';
import '../entities/handoff_task.dart';

abstract class DispatchRepository {
  Future<Result<List<HandoffTask>>> listQueue();
  Future<Result<HandoffTask>> getHandoff(String id);
  Future<Result<HandoffTask>> markCourierArrived(String id, {required String idempotencyKey});
  Future<Result<HandoffTask>> scanHandoff({
    required String id,
    required String scannedToken,
    String? scannedOrderId,
    required String idempotencyKey,
  });
  Future<Result<HandoffTask>> confirm(String id, {required String idempotencyKey});
  Future<Result<HandoffTask>> fail(String id, {required String reasonCode, String? notes, required String idempotencyKey});
}
