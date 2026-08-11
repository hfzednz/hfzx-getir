import 'package:nexora_core/nexora_core.dart';

import '../entities/duty_status.dart';

abstract class DutyRepository {
  Future<Result<DutyStatus>> getStatus();
  Future<Result<DutyStatus>> setStatus(DutyStatus status);
}
