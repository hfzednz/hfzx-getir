import 'package:nexora_core/nexora_core.dart';
import '../entities/shift_entity.dart';

abstract class ShiftsRepository {
  Future<Result<WarehouseShift?>> current();
  Future<Result<WarehouseShift>> clockIn({required String idempotencyKey});
  Future<Result<WarehouseShift>> clockOut({required String idempotencyKey});
  Future<Result<WarehouseShift>> startBreak({required String idempotencyKey});
  Future<Result<WarehouseShift>> endBreak({required String idempotencyKey});
  Future<Result<List<WarehouseShift>>> attendance();
}
