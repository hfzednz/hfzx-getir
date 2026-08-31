import 'package:equatable/equatable.dart';

enum SupportTicketStatus { open, inProgress, resolved, closed }

class SupportFaq extends Equatable {
  const SupportFaq({
    required this.id,
    required this.question,
    required this.answer,
    this.category = '',
  });

  final String id;
  final String question;
  final String answer;
  final String category;

  factory SupportFaq.fromJson(Map<String, dynamic> json) => SupportFaq(
        id: json['id']?.toString() ?? '',
        question: json['question']?.toString() ?? json['title']?.toString() ?? '',
        answer: json['answer']?.toString() ?? json['body']?.toString() ?? '',
        category: json['category']?.toString() ?? '',
      );

  @override
  List<Object?> get props => [id, question, answer];
}

class SupportTicketMessage extends Equatable {
  const SupportTicketMessage({
    required this.id,
    required this.body,
    this.author = 'support',
    this.createdAt,
    this.attachments = const [],
  });

  final String id;
  final String body;
  final String author;
  final DateTime? createdAt;
  final List<String> attachments;

  bool get isFromUser => author == 'user' || author == 'customer';

  factory SupportTicketMessage.fromJson(Map<String, dynamic> json) => SupportTicketMessage(
        id: json['id']?.toString() ?? '',
        body: json['body']?.toString() ?? json['message']?.toString() ?? '',
        author: json['author']?.toString() ?? 'support',
        createdAt: json['created_at'] != null ? DateTime.tryParse(json['created_at'].toString()) : null,
        attachments: (json['attachments'] as List<dynamic>?)?.map((e) => e.toString()).toList() ?? const [],
      );

  @override
  List<Object?> get props => [id, body, author, createdAt];
}

class SupportTicket extends Equatable {
  const SupportTicket({
    required this.id,
    required this.subject,
    this.status = SupportTicketStatus.open,
    this.orderId,
    this.category = '',
    this.messages = const [],
    this.createdAt,
    this.liveChatUrl,
  });

  final String id;
  final String subject;
  final SupportTicketStatus status;
  final String? orderId;
  final String category;
  final List<SupportTicketMessage> messages;
  final DateTime? createdAt;
  final String? liveChatUrl;

  factory SupportTicket.fromJson(Map<String, dynamic> json) {
    final messagesRaw = json['messages'] as List<dynamic>? ?? json['timeline'] as List<dynamic>?;
    return SupportTicket(
        id: json['id']?.toString() ?? json['ticketId']?.toString() ?? '',
      subject: json['subject']?.toString() ?? json['title']?.toString() ?? '',
      status: SupportTicketStatus.values.asNameMap()[json['status']?.toString()] ??
          SupportTicketStatus.open,
      orderId: json['order_id']?.toString(),
      category: json['category']?.toString() ?? '',
      messages: messagesRaw
              ?.map((e) => SupportTicketMessage.fromJson(e as Map<String, dynamic>))
              .toList() ??
          const [],
      createdAt: json['created_at'] != null ? DateTime.tryParse(json['created_at'].toString()) : null,
      liveChatUrl: json['live_chat_url']?.toString(),
    );
  }

  @override
  List<Object?> get props => [id, subject, status, orderId, messages.length];
}

class SupportAssistantMessage extends Equatable {
  const SupportAssistantMessage({
    required this.id,
    required this.role,
    required this.content,
    this.createdAt,
    this.suggestedActions = const [],
  });

  final String id;
  final String role;
  final String content;
  final DateTime? createdAt;
  final List<String> suggestedActions;

  bool get isAssistant => role == 'assistant';

  factory SupportAssistantMessage.fromJson(Map<String, dynamic> json) => SupportAssistantMessage(
        id: json['id']?.toString() ?? '',
        role: json['role']?.toString() ?? 'assistant',
        content: json['content']?.toString() ?? json['message']?.toString() ?? '',
        createdAt: json['created_at'] != null ? DateTime.tryParse(json['created_at'].toString()) : null,
        suggestedActions: (json['suggested_actions'] as List<dynamic>?)?.map((e) => e.toString()).toList() ??
            const [],
      );

  @override
  List<Object?> get props => [id, role, content];
}
