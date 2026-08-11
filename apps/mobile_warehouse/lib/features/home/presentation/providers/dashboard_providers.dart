import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../di/providers.dart';
import '../../data/datasources/dashboard_remote_datasource.dart';
import '../../data/repositories/dashboard_repository_impl.dart';
import '../../domain/entities/dashboard_entity.dart';
import '../../domain/repositories/dashboard_repository.dart';

final dashboardRemoteDataSourceProvider =
    Provider<DashboardRemoteDataSource>((ref) {
  return DashboardRemoteDataSource(ref.watch(apiClientProvider));
});

final dashboardRepositoryProvider = Provider<DashboardRepository>((ref) {
  return DashboardRepositoryImpl(ref.watch(dashboardRemoteDataSourceProvider));
});

final dashboardProvider =
    FutureProvider.autoDispose<WarehouseDashboard>((ref) async {
  final result = await ref.watch(dashboardRepositoryProvider).fetchDashboard();
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});
