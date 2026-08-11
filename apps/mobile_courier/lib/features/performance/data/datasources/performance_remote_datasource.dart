import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/performance_entity.dart';

class PerformanceRemoteDataSource {
  const PerformanceRemoteDataSource(this._client);
  final ApiClient _client;

  Future<Result<PerformanceMetrics>> fetch() {
    return _client.get<PerformanceMetrics>(
      '/courier/performance',
      parser: (json) =>
          PerformanceMetrics.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }
}
