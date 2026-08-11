import 'package:nexora_core/nexora_core.dart';
import '../../domain/entities/expiry_item.dart';

class ExpiryRemoteDataSource {
  const ExpiryRemoteDataSource(this._client);
  final ApiClient _client;

  Future<Result<List<ExpiryItem>>> listNearExpiry() {
    return _client.get<List<ExpiryItem>>(
      '/warehouse/expiry',
      parser: (json) {
        final list = json is List ? json : (json as Map)['items'] as List? ?? const [];
        return list.map((e) => ExpiryItem.fromJson(Map<String, dynamic>.from(e as Map))).toList();
      },
    );
  }

  Future<Result<void>> wasteRemove({required String sku, required int qty, required String idempotencyKey}) {
    return _client.post<void>(
      '/warehouse/expiry/waste',
      data: {'sku': sku, 'qty': qty},
      idempotencyKey: idempotencyKey,
    );
  }
}
