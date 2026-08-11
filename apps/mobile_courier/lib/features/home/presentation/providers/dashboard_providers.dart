import 'package:battery_plus/battery_plus.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../di/providers.dart';
import '../../../../shared/business_rules/location_rules.dart';
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
    FutureProvider.autoDispose<CourierDashboard>((ref) async {
  final result = await ref.watch(dashboardRepositoryProvider).fetchDashboard();
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});

final batteryLevelProvider = FutureProvider.autoDispose<int>((ref) async {
  try {
    final level = await Battery().batteryLevel;
    return level;
  } catch (_) {
    return 100;
  }
});

final batteryLowWarningProvider = Provider.autoDispose<bool>((ref) {
  final level = ref.watch(batteryLevelProvider).valueOrNull ?? 100;
  return LocationRules.isLowBattery(level);
});

final connectionQualityProvider = Provider.autoDispose<String>((ref) {
  final online = ref.watch(connectivityOnlineProvider).valueOrNull ?? true;
  if (!online) return 'offline';
  final dash = ref.watch(dashboardProvider).valueOrNull;
  return dash?.connectionQuality ?? 'good';
});
