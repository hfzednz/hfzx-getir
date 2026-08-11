import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/profile_entity.dart';
import '../models/profile_model.dart';

class ProfileRemoteDataSource {
  const ProfileRemoteDataSource(this._client);
  final ApiClient _client;

  static const _profilePath = '/profile';

  Future<Result<UserProfile>> fetchProfile() async {
    return _client.get<UserProfile>(
      _profilePath,
      parser: (json) => UserProfileModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  Future<Result<UserProfile>> updateProfile(UserProfile profile, {String? idempotencyKey}) async {
    return _client.patch<UserProfile>(
      _profilePath,
      data: profile.toJson(),
      idempotencyKey: idempotencyKey,
      parser: (json) => UserProfileModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }
}
