import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/loyalty_entity.dart';
import '../models/loyalty_model.dart';

class LoyaltyRemoteDataSource {
  const LoyaltyRemoteDataSource(this._client);
  final ApiClient _client;

  static const _accountPath = '/loyalty/account';
  static const _achievementsPath = '/loyalty/achievements';
  static const _badgesPath = '/loyalty/badges';
  static const _milestonesPath = '/loyalty/milestones';
  static const _dailyClaimPath = '/loyalty/rewards/daily/claim';

  Future<Result<LoyaltyAccount>> fetchAccount() async {
    return _client.get<LoyaltyAccount>(
      _accountPath,
      parser: (json) => LoyaltyAccountModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  Future<Result<List<LoyaltyAchievement>>> fetchAchievements() async {
    return _client.get<List<LoyaltyAchievement>>(
      _achievementsPath,
      parser: (json) => (json as List<dynamic>)
          .map((e) => LoyaltyAchievementModel.fromJson(e as Map<String, dynamic>).toEntity())
          .toList(),
    );
  }

  Future<Result<List<LoyaltyBadge>>> fetchBadges() async {
    return _client.get<List<LoyaltyBadge>>(
      _badgesPath,
      parser: (json) => (json as List<dynamic>)
          .map((e) => LoyaltyBadgeModel.fromJson(e as Map<String, dynamic>).toEntity())
          .toList(),
    );
  }

  Future<Result<List<LoyaltyMilestone>>> fetchMilestones() async {
    return _client.get<List<LoyaltyMilestone>>(
      _milestonesPath,
      parser: (json) => (json as List<dynamic>)
          .map((e) => LoyaltyMilestoneModel.fromJson(e as Map<String, dynamic>).toEntity())
          .toList(),
    );
  }

  Future<Result<DailyRewardClaim>> claimDailyReward({String? idempotencyKey}) async {
    return _client.post<DailyRewardClaim>(
      _dailyClaimPath,
      idempotencyKey: idempotencyKey,
      parser: (json) => DailyRewardClaimModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }
}
