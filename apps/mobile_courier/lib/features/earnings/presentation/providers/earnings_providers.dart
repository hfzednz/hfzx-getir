import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../di/providers.dart';
import '../../data/datasources/earnings_remote_datasource.dart';
import '../../data/repositories/earnings_repository_impl.dart';
import '../../domain/entities/earnings_entity.dart';
import '../../domain/repositories/earnings_repository.dart';

final earningsRemoteDataSourceProvider =
    Provider<EarningsRemoteDataSource>((ref) {
  return EarningsRemoteDataSource(ref.watch(apiClientProvider));
});

final earningsRepositoryProvider = Provider<EarningsRepository>((ref) {
  return EarningsRepositoryImpl(ref.watch(earningsRemoteDataSourceProvider));
});

final earningsPeriodProvider =
    StateProvider<EarningsPeriod>((ref) => EarningsPeriod.daily);

final earningsProvider =
    FutureProvider.autoDispose<EarningsSnapshot>((ref) async {
  final period = ref.watch(earningsPeriodProvider);
  final result = await ref.watch(earningsRepositoryProvider).fetch(period);
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});
