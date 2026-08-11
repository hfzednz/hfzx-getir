import 'package:equatable/equatable.dart';

enum ReferralClaimStatus { pending, approved, rejected, flagged, expired }

class ReferralDeviceMetadata extends Equatable {
  const ReferralDeviceMetadata({
    required this.deviceId,
    this.model,
    this.os,
    this.appVersion,
  });

  final String deviceId;
  final String? model;
  final String? os;
  final String? appVersion;

  Map<String, dynamic> toJson() => {
        'device_id': deviceId,
        if (model != null && model!.isNotEmpty) 'model': model,
        if (os != null && os!.isNotEmpty) 'os': os,
        if (appVersion != null && appVersion!.isNotEmpty) 'app_version': appVersion,
      };

  @override
  List<Object?> get props => [deviceId, model, os, appVersion];
}

class ReferralInfo extends Equatable {
  const ReferralInfo({
    required this.id,
    required this.inviteCode,
    this.shareUrl = '',
    this.rewardMinor = 0,
    this.currency = 'TRY',
    this.totalInvites = 0,
    this.successfulInvites = 0,
    this.pendingRewardMinor = 0,
  });

  final String id;
  final String inviteCode;
  final String shareUrl;
  final int rewardMinor;
  final String currency;
  final int totalInvites;
  final int successfulInvites;
  final int pendingRewardMinor;

  String get shareMessage => 'Join NEXORA with code $inviteCode';

  factory ReferralInfo.fromJson(Map<String, dynamic> json) => ReferralInfo(
        id: json['id']?.toString() ?? '',
        inviteCode: (json['invite_code']?.toString() ?? json['code']?.toString() ?? '').toUpperCase(),
        shareUrl: json['share_url']?.toString() ?? '',
        rewardMinor: (json['reward_minor'] as num?)?.toInt() ?? 0,
        currency: json['currency']?.toString() ?? 'TRY',
        totalInvites: (json['total_invites'] as num?)?.toInt() ?? 0,
        successfulInvites: (json['successful_invites'] as num?)?.toInt() ?? 0,
        pendingRewardMinor: (json['pending_reward_minor'] as num?)?.toInt() ?? 0,
      );

  @override
  List<Object?> get props => [id, inviteCode, totalInvites, successfulInvites];
}

class ReferralInvite extends Equatable {
  const ReferralInvite({
    required this.id,
    this.refereeLabel = '',
    this.claimStatus = ReferralClaimStatus.pending,
    this.rewardMinor = 0,
    this.currency = 'TRY',
    this.createdAt,
    this.fraudFlag = false,
    this.rejectedReason,
  });

  final String id;
  final String refereeLabel;
  final ReferralClaimStatus claimStatus;
  final int rewardMinor;
  final String currency;
  final DateTime? createdAt;
  final bool fraudFlag;
  final String? rejectedReason;

  factory ReferralInvite.fromJson(Map<String, dynamic> json) => ReferralInvite(
        id: json['id']?.toString() ?? '',
        refereeLabel: json['referee_label']?.toString() ?? json['name']?.toString() ?? 'Friend',
        claimStatus: ReferralClaimStatus.values.asNameMap()[json['claim_status']?.toString()] ??
            ReferralClaimStatus.pending,
        rewardMinor: (json['reward_minor'] as num?)?.toInt() ?? 0,
        currency: json['currency']?.toString() ?? 'TRY',
        createdAt: json['created_at'] != null ? DateTime.tryParse(json['created_at'].toString()) : null,
        fraudFlag: json['fraud_flag'] as bool? ??
            json['fraudFlag'] as bool? ??
            false,
        rejectedReason: json['rejected_reason']?.toString() ??
            json['reject_reason']?.toString() ??
            json['fraud_reason']?.toString() ??
            json['rejection_reason']?.toString(),
      );

  @override
  List<Object?> get props => [id, claimStatus, fraudFlag, rejectedReason, rewardMinor];
}
