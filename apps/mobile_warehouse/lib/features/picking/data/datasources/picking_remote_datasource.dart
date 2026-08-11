import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/pick_task.dart';

class PickingRemoteDataSource {
  const PickingRemoteDataSource(this._client);
  final ApiClient _client;

  Future<Result<List<PickTask>>> listQueue() {
    return _client.get<List<PickTask>>(
      '/warehouse/picking/queue',
      parser: (json) {
        final list = json is List
            ? json
            : (json as Map)['items'] as List? ?? const [];
        return list
            .map((e) => PickTask.fromJson(Map<String, dynamic>.from(e as Map)))
            .toList();
      },
    );
  }

  Future<Result<PickTask>> getTask(String taskId) {
    return _client.get<PickTask>(
      '/warehouse/picking/$taskId',
      parser: (json) =>
          PickTask.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }

  Future<Result<PickTask>> claim(String taskId, {required String idempotencyKey}) {
    return _client.post<PickTask>(
      '/warehouse/picking/$taskId/claim',
      idempotencyKey: idempotencyKey,
      parser: (json) =>
          PickTask.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }

  Future<Result<PickTask>> start(String taskId, {required String idempotencyKey}) {
    return _client.post<PickTask>(
      '/warehouse/picking/$taskId/start',
      idempotencyKey: idempotencyKey,
      parser: (json) =>
          PickTask.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }

  Future<Result<PickTask>> scanLine({
    required String taskId,
    required String lineId,
    required String barcode,
    required int qty,
    required String idempotencyKey,
  }) {
    return _client.post<PickTask>(
      '/warehouse/picking/$taskId/lines/$lineId/scan',
      data: {'barcode': barcode, 'qty': qty},
      idempotencyKey: idempotencyKey,
      parser: (json) =>
          PickTask.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }

  Future<Result<PickTask>> shortPick({
    required String taskId,
    required String lineId,
    required int missingQty,
    required String idempotencyKey,
  }) {
    return _client.post<PickTask>(
      '/warehouse/picking/$taskId/lines/$lineId/short-pick',
      data: {'missing_qty': missingQty},
      idempotencyKey: idempotencyKey,
      parser: (json) =>
          PickTask.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }

  Future<Result<PickTask>> complete(String taskId, {required String idempotencyKey}) {
    return _client.post<PickTask>(
      '/warehouse/picking/$taskId/complete',
      idempotencyKey: idempotencyKey,
      parser: (json) =>
          PickTask.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }

  Future<Result<PickTask>> stage(String taskId, {required String idempotencyKey}) {
    return _client.post<PickTask>(
      '/warehouse/picking/$taskId/stage',
      idempotencyKey: idempotencyKey,
      parser: (json) =>
          PickTask.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }
}
