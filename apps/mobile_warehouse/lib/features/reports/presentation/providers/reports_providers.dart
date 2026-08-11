import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../di/providers.dart';
import '../../data/datasources/reports_remote_datasource.dart';
import '../../data/repositories/reports_repository_impl.dart';
import '../../domain/entities/report_kpis.dart';
import '../../domain/repositories/reports_repository.dart';

final reportsRemoteDataSourceProvider = Provider((ref) => ReportsRemoteDataSource(ref.watch(apiClientProvider)));
final reportsRepositoryProvider = Provider<ReportsRepository>((ref) => ReportsRepositoryImpl(ref.watch(reportsRemoteDataSourceProvider)));
final reportKpisProvider = FutureProvider.autoDispose<ReportKpis>((ref) async {
  final r = await ref.watch(reportsRepositoryProvider).fetchKpis();
  return r.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});
