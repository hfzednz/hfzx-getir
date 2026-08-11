import '../../domain/entities/notifications_entity.dart';

class AppNotificationModel {
  const AppNotificationModel({required this.id, required this.title, required this.raw});

  final String id;
  final String title;
  final Map<String, dynamic> raw;

  factory AppNotificationModel.fromJson(Map<String, dynamic> json) =>
      AppNotificationModel(
        id: json['id']?.toString() ?? '',
        title: json['title']?.toString() ?? json['name']?.toString() ?? '',
        raw: json,
      );

  AppNotification toEntity() => AppNotification.fromJson(raw);
}
