import 'package:nexora_core/nexora_core.dart';

import '../entities/earnings_entity.dart';

abstract class EarningsRepository {
  Future<Result<EarningsSnapshot>> fetch(EarningsPeriod period);
}
