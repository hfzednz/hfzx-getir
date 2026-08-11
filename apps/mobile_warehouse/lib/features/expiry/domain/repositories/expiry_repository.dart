import 'package:nexora_core/nexora_core.dart';
import '../entities/expiry_item.dart';

abstract class ExpiryRepository {
  Future<Result<List<ExpiryItem>>> listNearExpiry();
  Future<Result<void>> wasteRemove({required String sku, required int qty, required String idempotencyKey});
}
