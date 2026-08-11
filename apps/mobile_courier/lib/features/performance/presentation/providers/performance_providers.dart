import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../di/providers.dart';
import '../../data/datasources/performance_remote_datasource.dart';
import '../../data/repositories/performance_repository_impl.dart';
import '../../domain/entities/performance_entity.dart';
import '../../domain/repositories/performance_repository.dart';

final performanceRemoteDataSourceProvider =
    Provider<PerformanceRemoteDataSource>((ref) {
  return PerformanceRemoteDataSource(ref.watch(apiClientProvider));
});

final performanceRepositoryProvider = Provider<PerformanceRepository>((ref) {
  return PerformanceRepositoryImpl(
    ref.watch(performanceRemoteDataSourceProvider),
  );
});

final performanceProvider =
    FutureProvider.autoDispose<PerformanceMetrics>((ref) async {
  final result = await ref.watch(performanceRepositoryProvider).fetch();
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});
