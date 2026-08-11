import '../../domain/entities/auth_entity.dart';

class AuthSessionModel {
  const AuthSessionModel({required this.id, required this.title, required this.raw});

  final String id;
  final String title;
  final Map<String, dynamic> raw;

  factory AuthSessionModel.fromJson(Map<String, dynamic> json) => AuthSessionModel(
        id: json['id']?.toString() ?? '',
        title: json['title']?.toString() ?? json['name']?.toString() ?? '',
        raw: json,
      );

  AuthSession toEntity() => AuthSession(id: id, title: title, payload: raw);
}
