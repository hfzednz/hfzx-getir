import 'package:equatable/equatable.dart';

class OpsTask extends Equatable {
  const OpsTask({
    required this.id,
    required this.title,
    required this.category,
    required this.status,
    this.priority = 0,
  });
  final String id;
  final String title;
  final String category; // cleaning | maintenance | emergency
  final String status;
  final int priority;
  factory OpsTask.fromJson(Map<String, dynamic> json) => OpsTask(
        id: json['id']?.toString() ?? '',
        title: json['title']?.toString() ?? '',
        category: json['category']?.toString() ?? 'cleaning',
        status: json['status']?.toString() ?? 'open',
        priority: (json['priority'] as num?)?.toInt() ?? 0,
      );
  @override
  List<Object?> get props => [id, title, category, status, priority];
}
