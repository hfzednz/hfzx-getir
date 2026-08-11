import 'package:nexora_core/nexora_core.dart';

import '../entities/loyalty_entity.dart';
import '../repositories/loyalty_repository.dart';

class GetLoyaltyAccountUseCase {
  const GetLoyaltyAccountUseCase(this._repository);
  final LoyaltyRepository _repository;
  Future<Result<LoyaltyAccount>> call() => _repository.fetchAccount();
}

class ListLoyaltyAchievementsUseCase {
  const ListLoyaltyAchievementsUseCase(this._repository);
  final LoyaltyRepository _repository;
  Future<Result<List<LoyaltyAchievement>>> call() => _repository.fetchAchievements();
}

class ClaimDailyRewardUseCase {
  const ClaimDailyRewardUseCase(this._repository);
  final LoyaltyRepository _repository;
  Future<Result<DailyRewardClaim>> call({String? idempotencyKey}) =>
      _repository.claimDailyReward(idempotencyKey: idempotencyKey);
}
