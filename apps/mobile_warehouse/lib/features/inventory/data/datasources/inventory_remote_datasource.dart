import 'package:nexora_core/nexora_core.dart';
import '../../domain/entities/stock_item.dart';

class InventoryRemoteDataSource {
  const InventoryRemoteDataSource(this._client);
  final ApiClient _client;

  Future<Result<List<StockItem>>> listStock({String? query}) {
    return _client.get<List<StockItem>>(
      '/warehouse/inventory/stock',
      queryParameters: {if (query != null && query.isNotEmpty) 'q': query},
      parser: (json) {
        final list = json is List ? json : (json as Map)['items'] as List? ?? const [];
        return list.map((e) => StockItem.fromJson(Map<String, dynamic>.from(e as Map))).toList();
      },
    );
  }

  Future<Result<StockItem>> adjust({
    required String sku,
    required int delta,
    required String reasonCode,
    String? notes,
    required String idempotencyKey,
  }) {
    return _client.post<StockItem>(
      '/warehouse/inventory/adjust',
      data: {
        'sku': sku,
        'delta': delta,
        'reason_code': reasonCode,
        if (notes != null) 'notes': notes,
      },
      idempotencyKey: idempotencyKey,
      parser: (json) => StockItem.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }

  Future<Result<CycleCountSession>> startCycleCount({required String idempotencyKey}) {
    return _client.post<CycleCountSession>(
      '/warehouse/inventory/cycle-count',
      idempotencyKey: idempotencyKey,
      parser: (json) => CycleCountSession.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }

  Future<Result<CycleCountSession>> submitCount({
    required String sessionId,
    required String sku,
    required int countedQty,
    required String idempotencyKey,
  }) {
    return _client.post<CycleCountSession>(
      '/warehouse/inventory/cycle-count/$sessionId/count',
      data: {'sku': sku, 'counted_qty': countedQty},
      idempotencyKey: idempotencyKey,
      parser: (json) => CycleCountSession.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }

  Future<Result<void>> receiveInbound({
    required String reference,
    required List<Map<String, dynamic>> lines,
    required String idempotencyKey,
  }) {
    return _client.post<void>(
      '/warehouse/inventory/inbound/receive',
      data: {'reference': reference, 'lines': lines},
      idempotencyKey: idempotencyKey,
    );
  }
}
