import 'package:equatable/equatable.dart';

class CourierNotification extends Equatable {
  const CourierNotification({
    required this.id,
    required this.title,
    required this.body,
    this.read = false,
    this.createdAt,
  });

  final String id;
  final String title;
  final String body;
  final bool read;
  final DateTime? createdAt;

  factory CourierNotification.fromJson(Map<String, dynamic> json) {
    return CourierNotification(
      id: json['id']?.toString() ?? '',
      title: json['title']?.toString() ?? '',
      body: json['body']?.toString() ?? json['message']?.toString() ?? '',
      read: json['read'] == true,
      createdAt: DateTime.tryParse(json['created_at']?.toString() ?? ''),
    );
  }

  @override
  List<Object?> get props => [id, title, body, read, createdAt];
}
