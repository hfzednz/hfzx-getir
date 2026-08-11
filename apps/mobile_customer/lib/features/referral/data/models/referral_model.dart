import '../../domain/entities/referral_entity.dart';

class ReferralInfoModel {
  const ReferralInfoModel({required this.raw});
  final Map<String, dynamic> raw;
  factory ReferralInfoModel.fromJson(Map<String, dynamic> json) => ReferralInfoModel(raw: json);
  ReferralInfo toEntity() => ReferralInfo.fromJson(raw);
}

class ReferralInviteModel {
  const ReferralInviteModel({required this.raw});
  final Map<String, dynamic> raw;
  factory ReferralInviteModel.fromJson(Map<String, dynamic> json) => ReferralInviteModel(raw: json);
  ReferralInvite toEntity() => ReferralInvite.fromJson(raw);
}
