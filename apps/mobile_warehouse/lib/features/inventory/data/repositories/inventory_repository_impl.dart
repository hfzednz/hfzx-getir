import 'package:nexora_core/nexora_core.dart';
import '../../domain/entities/stock_item.dart';
import '../../domain/repositories/inventory_repository.dart';
import '../datasources/inventory_remote_datasource.dart';

class InventoryRepositoryImpl implements InventoryRepository {
  InventoryRepositoryImpl(this._remote);
  final InventoryRemoteDataSource _remote;

  @override
  Future<Result<List<StockItem>>> listStock({String? query}) => _remote.listStock(query: query);

  @override
  Future<Result<StockItem>> adjust({
    required String sku,
    required int delta,
    required String reasonCode,
    String? notes,
    required String idempotencyKey,
  }) =>
      _remote.adjust(sku: sku, delta: delta, reasonCode: reasonCode, notes: notes, idempotencyKey: idempotencyKey);

  @override
  Future<Result<CycleCountSession>> startCycleCount({required String idempotencyKey}) =>
      _remote.startCycleCount(idempotencyKey: idempotencyKey);

  @override
  Future<Result<CycleCountSession>> submitCount({
    required String sessionId,
    required String sku,
    required int countedQty,
    required String idempotencyKey,
  }) =>
      _remote.submitCount(sessionId: sessionId, sku: sku, countedQty: countedQty, idempotencyKey: idempotencyKey);

  @override
  Future<Result<void>> receiveInbound({
    required String reference,
    required List<Map<String, dynamic>> lines,
    required String idempotencyKey,
  }) =>
      _remote.receiveInbound(reference: reference, lines: lines, idempotencyKey: idempotencyKey);
}
