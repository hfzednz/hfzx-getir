import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/duty_status.dart';

class DutyRemoteDataSource {
  const DutyRemoteDataSource(this._client);
  final ApiClient _client;

  Future<Result<DutyStatus>> fetchStatus() {
    return _client.get<DutyStatus>(
      '/courier/duty/status',
      parser: (json) {
        final map = json is Map<String, dynamic>
            ? json
            : Map<String, dynamic>.from(json as Map);
        return DutyStatus.fromApi(
          map['status']?.toString() ?? map['duty_status']?.toString(),
        );
      },
    );
  }

  Future<Result<DutyStatus>> postStatus(DutyStatus status) {
    return _client.post<DutyStatus>(
      '/courier/duty/status',
      data: {'status': status.apiValue},
      parser: (json) {
        final map = json is Map<String, dynamic>
            ? json
            : Map<String, dynamic>.from(json as Map);
        return DutyStatus.fromApi(
          map['status']?.toString() ?? status.apiValue,
        );
      },
    );
  }
}
