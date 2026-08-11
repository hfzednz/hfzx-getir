import 'package:nexora_core/nexora_core.dart';

import '../entities/shift_entity.dart';

abstract class ShiftsRepository {
  Future<Result<List<CourierShift>>> listShifts();
  Future<Result<CourierShift>> startShift();
  Future<Result<CourierShift>> endShift(String id);
  Future<Result<CourierShift>> startBreak(String id);
  Future<Result<CourierShift>> endBreak(String id);
}
