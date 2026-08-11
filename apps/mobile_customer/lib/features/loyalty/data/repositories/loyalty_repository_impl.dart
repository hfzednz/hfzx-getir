import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/loyalty_entity.dart';
import '../../domain/repositories/loyalty_repository.dart';
import '../datasources/loyalty_remote_datasource.dart';

class LoyaltyRepositoryImpl implements LoyaltyRepository {
  const LoyaltyRepositoryImpl(this._remote);
  final LoyaltyRemoteDataSource _remote;

  @override
  Future<Result<LoyaltyAccount>> fetchAccount() => _remote.fetchAccount();

  @override
  Future<Result<List<LoyaltyAchievement>>> fetchAchievements() => _remote.fetchAchievements();

  @override
  Future<Result<List<LoyaltyBadge>>> fetchBadges() => _remote.fetchBadges();

  @override
  Future<Result<List<LoyaltyMilestone>>> fetchMilestones() => _remote.fetchMilestones();

  @override
  Future<Result<DailyRewardClaim>> claimDailyReward({String? idempotencyKey}) =>
      _remote.claimDailyReward(idempotencyKey: idempotencyKey);
}
