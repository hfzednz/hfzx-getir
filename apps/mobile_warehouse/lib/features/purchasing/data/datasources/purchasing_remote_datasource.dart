import 'package:nexora_core/nexora_core.dart';
import '../../domain/entities/purchasing_entities.dart';

class PurchasingRemoteDataSource {
  const PurchasingRemoteDataSource(this._client);
  final ApiClient _client;

  Future<Result<List<Supplier>>> listSuppliers() {
    return _client.get<List<Supplier>>(
      '/warehouse/purchasing/suppliers',
      parser: (json) {
        final list = json is List ? json : (json as Map)['items'] as List? ?? const [];
        return list.map((e) => Supplier.fromJson(Map<String, dynamic>.from(e as Map))).toList();
      },
    );
  }

  Future<Result<List<PurchaseOrder>>> listPurchaseOrders() {
    return _client.get<List<PurchaseOrder>>(
      '/warehouse/purchasing/orders',
      parser: (json) {
        final list = json is List ? json : (json as Map)['items'] as List? ?? const [];
        return list.map((e) => PurchaseOrder.fromJson(Map<String, dynamic>.from(e as Map))).toList();
      },
    );
  }

  Future<Result<PurchaseOrder>> receivePo({
    required String poId,
    required bool qcFlag,
    required String idempotencyKey,
  }) {
    return _client.post<PurchaseOrder>(
      '/warehouse/purchasing/orders/$poId/receive',
      data: {'qc_flag': qcFlag},
      idempotencyKey: idempotencyKey,
      parser: (json) => PurchaseOrder.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }
}
