import 'package:equatable/equatable.dart';

class SupportTicket extends Equatable {
  const SupportTicket({
    required this.id,
    required this.subject,
    required this.status,
    this.updatedAt,
  });

  final String id;
  final String subject;
  final String status;
  final DateTime? updatedAt;

  factory SupportTicket.fromJson(Map<String, dynamic> json) {
    return SupportTicket(
      id: json['id']?.toString() ?? '',
      subject: json['subject']?.toString() ?? '',
      status: json['status']?.toString() ?? 'open',
      updatedAt: DateTime.tryParse(json['updated_at']?.toString() ?? ''),
    );
  }

  @override
  List<Object?> get props => [id, subject, status, updatedAt];
}

class ChatMessage extends Equatable {
  const ChatMessage({
    required this.id,
    required this.text,
    required this.fromAssistant,
    this.createdAt,
  });

  final String id;
  final String text;
  final bool fromAssistant;
  final DateTime? createdAt;

  factory ChatMessage.fromJson(Map<String, dynamic> json) {
    return ChatMessage(
      id: json['id']?.toString() ?? '',
      text: json['text']?.toString() ?? json['content']?.toString() ?? '',
      fromAssistant: json['from_assistant'] == true ||
          json['role']?.toString() == 'assistant',
      createdAt: DateTime.tryParse(json['created_at']?.toString() ?? ''),
    );
  }

  @override
  List<Object?> get props => [id, text, fromAssistant, createdAt];
}
