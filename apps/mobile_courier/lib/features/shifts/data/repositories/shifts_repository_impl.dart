import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/shift_entity.dart';
import '../../domain/repositories/shifts_repository.dart';
import '../datasources/shifts_remote_datasource.dart';

class ShiftsRepositoryImpl implements ShiftsRepository {
  ShiftsRepositoryImpl(this._remote);
  final ShiftsRemoteDataSource _remote;

  @override
  Future<Result<List<CourierShift>>> listShifts() => _remote.listShifts();

  @override
  Future<Result<CourierShift>> startShift() => _remote.postAction('/start');

  @override
  Future<Result<CourierShift>> endShift(String id) =>
      _remote.postAction('/end', id: id);

  @override
  Future<Result<CourierShift>> startBreak(String id) =>
      _remote.postAction('/break/start', id: id);

  @override
  Future<Result<CourierShift>> endBreak(String id) =>
      _remote.postAction('/break/end', id: id);
}
