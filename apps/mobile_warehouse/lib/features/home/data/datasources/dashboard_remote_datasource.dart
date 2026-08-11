import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/dashboard_entity.dart';

class DashboardRemoteDataSource {
  const DashboardRemoteDataSource(this._client);
  final ApiClient _client;

  Future<Result<WarehouseDashboard>> fetchDashboard() {
    return _client.get<WarehouseDashboard>(
      '/warehouse/dashboard',
      parser: (json) =>
          WarehouseDashboard.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }
}
