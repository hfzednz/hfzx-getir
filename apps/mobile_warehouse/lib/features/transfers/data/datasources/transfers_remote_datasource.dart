import 'package:nexora_core/nexora_core.dart';
import '../../domain/entities/transfer_entity.dart';

class TransfersRemoteDataSource {
  const TransfersRemoteDataSource(this._client);
  final ApiClient _client;

  Future<Result<List<WarehouseTransfer>>> list() {
    return _client.get<List<WarehouseTransfer>>(
      '/warehouse/transfers',
      parser: (json) {
        final list = json is List ? json : (json as Map)['items'] as List? ?? const [];
        return list.map((e) => WarehouseTransfer.fromJson(Map<String, dynamic>.from(e as Map))).toList();
      },
    );
  }

  Future<Result<WarehouseTransfer>> create(Map<String, dynamic> payload, {required String idempotencyKey}) {
    return _client.post<WarehouseTransfer>(
      '/warehouse/transfers',
      data: payload,
      idempotencyKey: idempotencyKey,
      parser: (json) => WarehouseTransfer.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }

  Future<Result<WarehouseTransfer>> approve(String id, {required String idempotencyKey}) {
    return _client.post<WarehouseTransfer>(
      '/warehouse/transfers/$id/approve',
      idempotencyKey: idempotencyKey,
      parser: (json) => WarehouseTransfer.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }
}
