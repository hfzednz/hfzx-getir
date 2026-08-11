import '../../domain/entities/profile_entity.dart';

class UserProfileModel {
  const UserProfileModel({required this.raw});
  final Map<String, dynamic> raw;
  factory UserProfileModel.fromJson(Map<String, dynamic> json) => UserProfileModel(raw: json);
  UserProfile toEntity() => UserProfile.fromJson(raw);
}
