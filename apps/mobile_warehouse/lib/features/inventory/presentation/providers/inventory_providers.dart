import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_core/nexora_core.dart';
import 'package:uuid/uuid.dart';

import '../../../../di/providers.dart';
import '../../../../shared/business_rules/inventory_rules.dart';
import '../../data/datasources/inventory_remote_datasource.dart';
import '../../data/repositories/inventory_repository_impl.dart';
import '../../domain/entities/stock_item.dart';
import '../../domain/repositories/inventory_repository.dart';

final inventoryRemoteDataSourceProvider = Provider<InventoryRemoteDataSource>((ref) {
  return InventoryRemoteDataSource(ref.watch(apiClientProvider));
});

final inventoryRepositoryProvider = Provider<InventoryRepository>((ref) {
  return InventoryRepositoryImpl(ref.watch(inventoryRemoteDataSourceProvider));
});

final stockListProvider = FutureProvider.autoDispose<List<StockItem>>((ref) async {
  final result = await ref.watch(inventoryRepositoryProvider).listStock();
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});

final inventoryActionsProvider = Provider((ref) => InventoryActions(ref));

class InventoryActions {
  InventoryActions(this._ref);
  final Ref _ref;
  final _uuid = const Uuid();

  Future<Result<StockItem>> adjust({
    required StockItem item,
    required int delta,
    required String reasonCode,
    String? notes,
  }) async {
    final v = InventoryRules.validateAdjust(
      delta: delta,
      reasonCode: reasonCode,
      notes: notes,
      currentOnHand: item.onHand,
    );
    if (v.isFailure) return Failure(v.errorOrNull!);
    final r = await _ref.read(inventoryRepositoryProvider).adjust(
          sku: item.sku,
          delta: delta,
          reasonCode: reasonCode,
          notes: notes,
          idempotencyKey: _uuid.v4(),
        );
    if (r.isSuccess) _ref.invalidate(stockListProvider);
    return r;
  }

  Future<Result<CycleCountSession>> startCycleCount() async {
    return _ref.read(inventoryRepositoryProvider).startCycleCount(idempotencyKey: _uuid.v4());
  }

  Future<Result<void>> receiveInbound({
    required String reference,
    required List<Map<String, dynamic>> lines,
  }) {
    return _ref.read(inventoryRepositoryProvider).receiveInbound(
          reference: reference,
          lines: lines,
          idempotencyKey: _uuid.v4(),
        );
  }
}
