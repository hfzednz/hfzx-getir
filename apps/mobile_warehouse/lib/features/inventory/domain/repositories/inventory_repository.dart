import 'package:nexora_core/nexora_core.dart';
import '../entities/stock_item.dart';

abstract class InventoryRepository {
  Future<Result<List<StockItem>>> listStock({String? query});
  Future<Result<StockItem>> adjust({
    required String sku,
    required int delta,
    required String reasonCode,
    String? notes,
    required String idempotencyKey,
  });
  Future<Result<CycleCountSession>> startCycleCount({required String idempotencyKey});
  Future<Result<CycleCountSession>> submitCount({
    required String sessionId,
    required String sku,
    required int countedQty,
    required String idempotencyKey,
  });
  Future<Result<void>> receiveInbound({
    required String reference,
    required List<Map<String, dynamic>> lines,
    required String idempotencyKey,
  });
}
