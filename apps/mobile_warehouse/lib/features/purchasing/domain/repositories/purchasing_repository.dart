import 'package:nexora_core/nexora_core.dart';
import '../entities/purchasing_entities.dart';

abstract class PurchasingRepository {
  Future<Result<List<Supplier>>> listSuppliers();
  Future<Result<List<PurchaseOrder>>> listPurchaseOrders();
  Future<Result<PurchaseOrder>> receivePo({
    required String poId,
    required bool qcFlag,
    required String idempotencyKey,
  });
}
