import 'package:equatable/equatable.dart';

class SupportTicket extends Equatable {
  const SupportTicket({required this.id, required this.subject, required this.status});
  final String id;
  final String subject;
  final String status;
  factory SupportTicket.fromJson(Map<String, dynamic> json) => SupportTicket(
        id: json['id']?.toString() ?? '',
        subject: json['subject']?.toString() ?? '',
        status: json['status']?.toString() ?? 'open',
      );
  @override
  List<Object?> get props => [id, subject, status];
}
