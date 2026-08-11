import 'package:nexora_core/nexora_core.dart';

import '../entities/loyalty_entity.dart';

abstract class LoyaltyRepository {
  Future<Result<LoyaltyAccount>> fetchAccount();
  Future<Result<List<LoyaltyAchievement>>> fetchAchievements();
  Future<Result<List<LoyaltyBadge>>> fetchBadges();
  Future<Result<List<LoyaltyMilestone>>> fetchMilestones();
  Future<Result<DailyRewardClaim>> claimDailyReward({String? idempotencyKey});
}
