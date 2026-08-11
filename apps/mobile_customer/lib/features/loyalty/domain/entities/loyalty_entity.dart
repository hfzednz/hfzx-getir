import 'package:equatable/equatable.dart';

enum LoyaltyTier { bronze, silver, gold, platinum, vip }

class LoyaltyAccount extends Equatable {
  const LoyaltyAccount({
    required this.id,
    this.points = 0,
    this.lifetimePoints = 0,
    this.tier = LoyaltyTier.bronze,
    this.tierLabel = 'Bronze',
    this.nextTierPoints = 0,
    this.tierProgressPercent = 0,
    this.dailyRewardAvailable = false,
    this.dailyRewardClaimedToday = false,
  });

  final String id;
  final int points;
  final int lifetimePoints;
  final LoyaltyTier tier;
  final String tierLabel;
  final int nextTierPoints;
  final double tierProgressPercent;
  final bool dailyRewardAvailable;
  final bool dailyRewardClaimedToday;

  factory LoyaltyAccount.fromJson(Map<String, dynamic> json) => LoyaltyAccount(
        id: json['id']?.toString() ?? '',
        points: (json['points'] as num?)?.toInt() ?? 0,
        lifetimePoints: (json['lifetime_points'] as num?)?.toInt() ?? 0,
        tier: LoyaltyTier.values.asNameMap()[json['tier']?.toString()] ?? LoyaltyTier.bronze,
        tierLabel: json['tier_label']?.toString() ?? json['tier']?.toString() ?? 'Bronze',
        nextTierPoints: (json['next_tier_points'] as num?)?.toInt() ?? 0,
        tierProgressPercent: (json['tier_progress_percent'] as num?)?.toDouble() ?? 0,
        dailyRewardAvailable: json['daily_reward_available'] as bool? ?? false,
        dailyRewardClaimedToday: json['daily_reward_claimed_today'] as bool? ?? false,
      );

  @override
  List<Object?> get props =>
      [id, points, tier, tierProgressPercent, dailyRewardAvailable, dailyRewardClaimedToday];
}

class LoyaltyAchievement extends Equatable {
  const LoyaltyAchievement({
    required this.id,
    required this.title,
    this.description = '',
    this.iconUrl,
    this.unlocked = false,
    this.unlockedAt,
    this.progressPercent = 0,
  });

  final String id;
  final String title;
  final String description;
  final String? iconUrl;
  final bool unlocked;
  final DateTime? unlockedAt;
  final double progressPercent;

  factory LoyaltyAchievement.fromJson(Map<String, dynamic> json) => LoyaltyAchievement(
        id: json['id']?.toString() ?? '',
        title: json['title']?.toString() ?? '',
        description: json['description']?.toString() ?? '',
        iconUrl: json['icon_url']?.toString(),
        unlocked: json['unlocked'] as bool? ?? false,
        unlockedAt:
            json['unlocked_at'] != null ? DateTime.tryParse(json['unlocked_at'].toString()) : null,
        progressPercent: (json['progress_percent'] as num?)?.toDouble() ?? 0,
      );

  @override
  List<Object?> get props => [id, title, unlocked, progressPercent];
}

class LoyaltyBadge extends Equatable {
  const LoyaltyBadge({
    required this.id,
    required this.label,
    this.imageUrl,
    this.earnedAt,
  });

  final String id;
  final String label;
  final String? imageUrl;
  final DateTime? earnedAt;

  factory LoyaltyBadge.fromJson(Map<String, dynamic> json) => LoyaltyBadge(
        id: json['id']?.toString() ?? '',
        label: json['label']?.toString() ?? json['title']?.toString() ?? '',
        imageUrl: json['image_url']?.toString(),
        earnedAt: json['earned_at'] != null ? DateTime.tryParse(json['earned_at'].toString()) : null,
      );

  @override
  List<Object?> get props => [id, label, earnedAt];
}

class LoyaltyMilestone extends Equatable {
  const LoyaltyMilestone({
    required this.id,
    required this.title,
    required this.targetPoints,
    this.rewardLabel = '',
    this.reached = false,
  });

  final String id;
  final String title;
  final int targetPoints;
  final String rewardLabel;
  final bool reached;

  factory LoyaltyMilestone.fromJson(Map<String, dynamic> json) => LoyaltyMilestone(
        id: json['id']?.toString() ?? '',
        title: json['title']?.toString() ?? '',
        targetPoints: (json['target_points'] as num?)?.toInt() ?? 0,
        rewardLabel: json['reward_label']?.toString() ?? '',
        reached: json['reached'] as bool? ?? false,
      );

  @override
  List<Object?> get props => [id, title, targetPoints, reached];
}

class DailyRewardClaim extends Equatable {
  const DailyRewardClaim({
    required this.pointsAwarded,
    this.streakDays = 1,
    this.message = '',
  });

  final int pointsAwarded;
  final int streakDays;
  final String message;

  factory DailyRewardClaim.fromJson(Map<String, dynamic> json) => DailyRewardClaim(
        pointsAwarded: (json['points_awarded'] as num?)?.toInt() ?? 0,
        streakDays: (json['streak_days'] as num?)?.toInt() ?? 1,
        message: json['message']?.toString() ?? '',
      );

  @override
  List<Object?> get props => [pointsAwarded, streakDays, message];
}
