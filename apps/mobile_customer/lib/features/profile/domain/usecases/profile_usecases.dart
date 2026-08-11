import 'package:nexora_core/nexora_core.dart';

import '../entities/profile_entity.dart';
import '../repositories/profile_repository.dart';

class GetUserProfileUseCase {
  const GetUserProfileUseCase(this._repository);
  final ProfileRepository _repository;

  Future<Result<UserProfile>> call() => _repository.fetchProfile();
}

class UpdateUserProfileUseCase {
  const UpdateUserProfileUseCase(this._repository);
  final ProfileRepository _repository;

  Future<Result<UserProfile>> call(UserProfile profile, {String? idempotencyKey}) =>
      _repository.updateProfile(profile, idempotencyKey: idempotencyKey);
}
