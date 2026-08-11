import 'package:nexora_core/nexora_core.dart';

import '../entities/referral_entity.dart';
import '../repositories/referral_repository.dart';

class GetReferralInfoUseCase {
  const GetReferralInfoUseCase(this._repository);
  final ReferralRepository _repository;

  Future<Result<ReferralInfo>> call() => _repository.fetchInfo();
}

class ListReferralInvitesUseCase {
  const ListReferralInvitesUseCase(this._repository);
  final ReferralRepository _repository;

  Future<Result<List<ReferralInvite>>> call() => _repository.fetchInvites();
}
