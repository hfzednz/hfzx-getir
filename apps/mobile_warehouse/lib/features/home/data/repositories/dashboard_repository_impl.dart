import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/dashboard_entity.dart';
import '../../domain/repositories/dashboard_repository.dart';
import '../datasources/dashboard_remote_datasource.dart';

class DashboardRepositoryImpl implements DashboardRepository {
  DashboardRepositoryImpl(this._remote);
  final DashboardRemoteDataSource _remote;

  @override
  Future<Result<WarehouseDashboard>> fetchDashboard() =>
      _remote.fetchDashboard();
}
