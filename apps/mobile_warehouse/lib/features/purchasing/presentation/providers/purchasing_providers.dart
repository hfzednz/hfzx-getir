import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_core/nexora_core.dart';
import 'package:uuid/uuid.dart';
import '../../../../di/providers.dart';
import '../../data/datasources/purchasing_remote_datasource.dart';
import '../../data/repositories/purchasing_repository_impl.dart';
import '../../domain/entities/purchasing_entities.dart';
import '../../domain/repositories/purchasing_repository.dart';

final purchasingRemoteDataSourceProvider = Provider((ref) => PurchasingRemoteDataSource(ref.watch(apiClientProvider)));
final purchasingRepositoryProvider = Provider<PurchasingRepository>((ref) => PurchasingRepositoryImpl(ref.watch(purchasingRemoteDataSourceProvider)));
final suppliersProvider = FutureProvider.autoDispose<List<Supplier>>((ref) async {
  final r = await ref.watch(purchasingRepositoryProvider).listSuppliers();
  return r.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});
final purchaseOrdersProvider = FutureProvider.autoDispose<List<PurchaseOrder>>((ref) async {
  final r = await ref.watch(purchasingRepositoryProvider).listPurchaseOrders();
  return r.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});
final purchasingActionsProvider = Provider((ref) => PurchasingActions(ref));

class PurchasingActions {
  PurchasingActions(this._ref);
  final Ref _ref;
  final _uuid = const Uuid();
  Future<Result<PurchaseOrder>> receivePo(String poId, {bool qcFlag = false}) async {
    final r = await _ref.read(purchasingRepositoryProvider).receivePo(poId: poId, qcFlag: qcFlag, idempotencyKey: _uuid.v4());
    if (r.isSuccess) _ref.invalidate(purchaseOrdersProvider);
    return r;
  }
}
