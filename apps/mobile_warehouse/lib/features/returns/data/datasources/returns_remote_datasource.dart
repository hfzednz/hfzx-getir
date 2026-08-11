import 'package:nexora_core/nexora_core.dart';
import '../../domain/entities/return_task.dart';

class ReturnsRemoteDataSource {
  const ReturnsRemoteDataSource(this._client);
  final ApiClient _client;

  Future<Result<List<ReturnTask>>> list({String? type}) {
    return _client.get<List<ReturnTask>>(
      '/warehouse/returns',
      queryParameters: {if (type != null) 'type': type},
      parser: (json) {
        final list = json is List ? json : (json as Map)['items'] as List? ?? const [];
        return list.map((e) => ReturnTask.fromJson(Map<String, dynamic>.from(e as Map))).toList();
      },
    );
  }

  Future<Result<ReturnTask>> advance(String id, {required String action, required String idempotencyKey}) {
    return _client.post<ReturnTask>(
      '/warehouse/returns/$id/$action',
      idempotencyKey: idempotencyKey,
      parser: (json) => ReturnTask.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }
}
