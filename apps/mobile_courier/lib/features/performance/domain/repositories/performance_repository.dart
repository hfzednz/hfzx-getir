import 'package:nexora_core/nexora_core.dart';

import '../entities/performance_entity.dart';

abstract class PerformanceRepository {
  Future<Result<PerformanceMetrics>> fetch();
}
