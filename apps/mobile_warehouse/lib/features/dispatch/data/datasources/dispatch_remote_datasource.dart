import 'package:nexora_core/nexora_core.dart';
import '../../domain/entities/handoff_task.dart';

class DispatchRemoteDataSource {
  const DispatchRemoteDataSource(this._client);
  final ApiClient _client;

  Future<Result<List<HandoffTask>>> listQueue() {
    return _client.get<List<HandoffTask>>(
      '/warehouse/dispatch/queue',
      parser: (json) {
        final list = json is List ? json : (json as Map)['items'] as List? ?? const [];
        return list.map((e) => HandoffTask.fromJson(Map<String, dynamic>.from(e as Map))).toList();
      },
    );
  }

  Future<Result<HandoffTask>> getHandoff(String id) {
    return _client.get<HandoffTask>(
      '/warehouse/dispatch/$id',
      parser: (json) => HandoffTask.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }

  Future<Result<HandoffTask>> markCourierArrived(String id, {required String idempotencyKey}) {
    return _client.post<HandoffTask>(
      '/warehouse/dispatch/$id/courier-arrived',
      idempotencyKey: idempotencyKey,
      parser: (json) => HandoffTask.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }

  Future<Result<HandoffTask>> scanHandoff({
    required String id,
    required String scannedToken,
    String? scannedOrderId,
    required String idempotencyKey,
  }) {
    return _client.post<HandoffTask>(
      '/warehouse/dispatch/$id/scan',
      data: {
        'scanned_token': scannedToken,
        if (scannedOrderId != null) 'scanned_order_id': scannedOrderId,
      },
      idempotencyKey: idempotencyKey,
      parser: (json) => HandoffTask.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }

  Future<Result<HandoffTask>> confirm(String id, {required String idempotencyKey}) {
    return _client.post<HandoffTask>(
      '/warehouse/dispatch/$id/confirm',
      idempotencyKey: idempotencyKey,
      parser: (json) => HandoffTask.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }

  Future<Result<HandoffTask>> fail(String id, {required String reasonCode, String? notes, required String idempotencyKey}) {
    return _client.post<HandoffTask>(
      '/warehouse/dispatch/$id/fail',
      data: {'reason_code': reasonCode, if (notes != null) 'notes': notes},
      idempotencyKey: idempotencyKey,
      parser: (json) => HandoffTask.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }
}
