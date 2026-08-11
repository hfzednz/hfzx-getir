import '../../domain/entities/settings_entity.dart';

class AppSettingsModel {
  const AppSettingsModel({required this.id, required this.title, required this.raw});

  final String id;
  final String title;
  final Map<String, dynamic> raw;

  factory AppSettingsModel.fromJson(Map<String, dynamic> json) => AppSettingsModel(
        id: json['id']?.toString() ?? '',
        title: json['title']?.toString() ?? json['name']?.toString() ?? '',
        raw: json,
      );

  AppSettings toEntity() => AppSettings(id: id, title: title, payload: raw);
}
