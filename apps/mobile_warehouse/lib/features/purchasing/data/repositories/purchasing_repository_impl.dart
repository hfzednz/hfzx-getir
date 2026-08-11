import 'package:nexora_core/nexora_core.dart';
import '../../domain/entities/purchasing_entities.dart';
import '../../domain/repositories/purchasing_repository.dart';
import '../datasources/purchasing_remote_datasource.dart';

class PurchasingRepositoryImpl implements PurchasingRepository {
  PurchasingRepositoryImpl(this._remote);
  final PurchasingRemoteDataSource _remote;
  @override
  Future<Result<List<Supplier>>> listSuppliers() => _remote.listSuppliers();
  @override
  Future<Result<List<PurchaseOrder>>> listPurchaseOrders() => _remote.listPurchaseOrders();
  @override
  Future<Result<PurchaseOrder>> receivePo({required String poId, required bool qcFlag, required String idempotencyKey}) =>
      _remote.receivePo(poId: poId, qcFlag: qcFlag, idempotencyKey: idempotencyKey);
}
