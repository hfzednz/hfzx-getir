import 'package:nexora_core/nexora_core.dart';
import '../../domain/entities/shift_entity.dart';
import '../../domain/repositories/shifts_repository.dart';
import '../datasources/shifts_remote_datasource.dart';

class ShiftsRepositoryImpl implements ShiftsRepository {
  ShiftsRepositoryImpl(this._remote);
  final ShiftsRemoteDataSource _remote;
  @override
  Future<Result<WarehouseShift?>> current() => _remote.current();
  @override
  Future<Result<WarehouseShift>> clockIn({required String idempotencyKey}) => _remote.clockIn(idempotencyKey: idempotencyKey);
  @override
  Future<Result<WarehouseShift>> clockOut({required String idempotencyKey}) => _remote.clockOut(idempotencyKey: idempotencyKey);
  @override
  Future<Result<WarehouseShift>> startBreak({required String idempotencyKey}) => _remote.startBreak(idempotencyKey: idempotencyKey);
  @override
  Future<Result<WarehouseShift>> endBreak({required String idempotencyKey}) => _remote.endBreak(idempotencyKey: idempotencyKey);
  @override
  Future<Result<List<WarehouseShift>>> attendance() => _remote.attendance();
}
