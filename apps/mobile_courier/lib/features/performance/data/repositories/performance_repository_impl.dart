import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/performance_entity.dart';
import '../../domain/repositories/performance_repository.dart';
import '../datasources/performance_remote_datasource.dart';

class PerformanceRepositoryImpl implements PerformanceRepository {
  PerformanceRepositoryImpl(this._remote);
  final PerformanceRemoteDataSource _remote;

  @override
  Future<Result<PerformanceMetrics>> fetch() => _remote.fetch();
}
