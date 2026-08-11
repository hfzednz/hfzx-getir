import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/profile_entity.dart';

class ProfileRemoteDataSource {
  const ProfileRemoteDataSource(this._client);
  final ApiClient _client;

  Future<Result<CourierProfile>> fetch() {
    return _client.get<CourierProfile>(
      '/courier/profile',
      parser: (json) =>
          CourierProfile.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }
}
