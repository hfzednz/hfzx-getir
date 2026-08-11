import 'package:equatable/equatable.dart';

class WarehouseNotification extends Equatable {
  const WarehouseNotification({
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
  factory WarehouseNotification.fromJson(Map<String, dynamic> json) => WarehouseNotification(
        id: json['id']?.toString() ?? '',
        title: json['title']?.toString() ?? '',
        body: json['body']?.toString() ?? '',
        read: json['read'] as bool? ?? false,
        createdAt: DateTime.tryParse(json['created_at']?.toString() ?? ''),
      );
  @override
  List<Object?> get props => [id, title, body, read];
}
