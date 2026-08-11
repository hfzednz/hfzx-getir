import 'package:nexora_core/nexora_core.dart';

import '../entities/dashboard_entity.dart';

abstract class DashboardRepository {
  Future<Result<WarehouseDashboard>> fetchDashboard();
}
