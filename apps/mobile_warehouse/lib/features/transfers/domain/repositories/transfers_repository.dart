import 'package:nexora_core/nexora_core.dart';
import '../entities/transfer_entity.dart';

abstract class TransfersRepository {
  Future<Result<List<WarehouseTransfer>>> list();
  Future<Result<WarehouseTransfer>> create(Map<String, dynamic> payload, {required String idempotencyKey});
  Future<Result<WarehouseTransfer>> approve(String id, {required String idempotencyKey});
}
