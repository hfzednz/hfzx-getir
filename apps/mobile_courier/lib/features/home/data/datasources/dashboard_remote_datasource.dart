import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/dashboard_entity.dart';

class DashboardRemoteDataSource {
  const DashboardRemoteDataSource(this._client);
  final ApiClient _client;

  Future<Result<CourierDashboard>> fetchDashboard() {
    return _client.get<CourierDashboard>(
      '/courier/dashboard',
      parser: (json) =>
          CourierDashboard.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }
}
