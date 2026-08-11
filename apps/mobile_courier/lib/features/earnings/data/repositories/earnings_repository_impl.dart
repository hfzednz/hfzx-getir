import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/earnings_entity.dart';
import '../../domain/repositories/earnings_repository.dart';
import '../datasources/earnings_remote_datasource.dart';

class EarningsRepositoryImpl implements EarningsRepository {
  EarningsRepositoryImpl(this._remote);
  final EarningsRemoteDataSource _remote;

  @override
  Future<Result<EarningsSnapshot>> fetch(EarningsPeriod period) =>
      _remote.fetch(period);
}
