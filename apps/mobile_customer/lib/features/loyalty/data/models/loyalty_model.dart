import '../../domain/entities/loyalty_entity.dart';

class LoyaltyAccountModel {
  const LoyaltyAccountModel({required this.raw});
  final Map<String, dynamic> raw;
  factory LoyaltyAccountModel.fromJson(Map<String, dynamic> json) => LoyaltyAccountModel(raw: json);
  LoyaltyAccount toEntity() => LoyaltyAccount.fromJson(raw);
}

class LoyaltyAchievementModel {
  const LoyaltyAchievementModel({required this.raw});
  final Map<String, dynamic> raw;
  factory LoyaltyAchievementModel.fromJson(Map<String, dynamic> json) =>
      LoyaltyAchievementModel(raw: json);
  LoyaltyAchievement toEntity() => LoyaltyAchievement.fromJson(raw);
}

class LoyaltyBadgeModel {
  const LoyaltyBadgeModel({required this.raw});
  final Map<String, dynamic> raw;
  factory LoyaltyBadgeModel.fromJson(Map<String, dynamic> json) => LoyaltyBadgeModel(raw: json);
  LoyaltyBadge toEntity() => LoyaltyBadge.fromJson(raw);
}

class LoyaltyMilestoneModel {
  const LoyaltyMilestoneModel({required this.raw});
  final Map<String, dynamic> raw;
  factory LoyaltyMilestoneModel.fromJson(Map<String, dynamic> json) => LoyaltyMilestoneModel(raw: json);
  LoyaltyMilestone toEntity() => LoyaltyMilestone.fromJson(raw);
}

class DailyRewardClaimModel {
  const DailyRewardClaimModel({required this.raw});
  final Map<String, dynamic> raw;
  factory DailyRewardClaimModel.fromJson(Map<String, dynamic> json) => DailyRewardClaimModel(raw: json);
  DailyRewardClaim toEntity() => DailyRewardClaim.fromJson(raw);
}
