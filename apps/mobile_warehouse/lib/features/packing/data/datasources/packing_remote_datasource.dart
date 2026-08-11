import 'package:nexora_core/nexora_core.dart';
import '../../domain/entities/pack_task.dart';

class PackingRemoteDataSource {
  const PackingRemoteDataSource(this._client);
  final ApiClient _client;

  Future<Result<List<PackTask>>> listQueue() {
    return _client.get<List<PackTask>>(
      '/warehouse/packing/queue',
      parser: (json) {
        final list = json is List ? json : (json as Map)['items'] as List? ?? const [];
        return list.map((e) => PackTask.fromJson(Map<String, dynamic>.from(e as Map))).toList();
      },
    );
  }

  Future<Result<PackTask>> getTask(String taskId) {
    return _client.get<PackTask>(
      '/warehouse/packing/$taskId',
      parser: (json) => PackTask.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }

  Future<Result<PackTask>> claim(String taskId, {required String idempotencyKey}) {
    return _client.post<PackTask>(
      '/warehouse/packing/$taskId/claim',
      idempotencyKey: idempotencyKey,
      parser: (json) => PackTask.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }

  Future<Result<PackTask>> weigh(String taskId, {required double actualGrams, required String idempotencyKey}) {
    return _client.post<PackTask>(
      '/warehouse/packing/$taskId/weigh',
      data: {'actual_weight_grams': actualGrams},
      idempotencyKey: idempotencyKey,
      parser: (json) => PackTask.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }

  Future<Result<PackTask>> printLabel(String taskId, {required String idempotencyKey}) {
    return _client.post<PackTask>(
      '/warehouse/packing/$taskId/label',
      idempotencyKey: idempotencyKey,
      parser: (json) => PackTask.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }

  Future<Result<PackTask>> seal(String taskId, {required String idempotencyKey}) {
    return _client.post<PackTask>(
      '/warehouse/packing/$taskId/seal',
      data: {'sealed': true},
      idempotencyKey: idempotencyKey,
      parser: (json) => PackTask.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }

  Future<Result<PackTask>> complete(String taskId, {required String idempotencyKey}) {
    return _client.post<PackTask>(
      '/warehouse/packing/$taskId/complete',
      idempotencyKey: idempotencyKey,
      parser: (json) => PackTask.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }
}
