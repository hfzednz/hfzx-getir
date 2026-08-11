import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/earnings_entity.dart';

class EarningsRemoteDataSource {
  const EarningsRemoteDataSource(this._client);
  final ApiClient _client;

  Future<Result<EarningsSnapshot>> fetch(EarningsPeriod period) {
    final periodParam = switch (period) {
      EarningsPeriod.daily => 'daily',
      EarningsPeriod.weekly => 'weekly',
      EarningsPeriod.monthly => 'monthly',
    };
    return _client.get<EarningsSnapshot>(
      '/courier/earnings',
      queryParameters: {'period': periodParam},
      parser: (json) => EarningsSnapshot.fromJson(
        Map<String, dynamic>.from(json as Map),
        period: period,
      ),
    );
  }
}
