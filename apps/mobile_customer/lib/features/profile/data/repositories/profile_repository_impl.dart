import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/profile_entity.dart';
import '../../domain/repositories/profile_repository.dart';
import '../datasources/profile_remote_datasource.dart';

class ProfileRepositoryImpl implements ProfileRepository {
  const ProfileRepositoryImpl(this._remote);
  final ProfileRemoteDataSource _remote;

  @override
  Future<Result<UserProfile>> fetchProfile() => _remote.fetchProfile();

  @override
  Future<Result<UserProfile>> updateProfile(UserProfile profile, {String? idempotencyKey}) =>
      _remote.updateProfile(profile, idempotencyKey: idempotencyKey);
}
