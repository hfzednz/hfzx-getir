import '../../domain/entities/support_entity.dart';

class SupportFaqModel {
  const SupportFaqModel({required this.raw});
  final Map<String, dynamic> raw;
  factory SupportFaqModel.fromJson(Map<String, dynamic> json) => SupportFaqModel(raw: json);
  SupportFaq toEntity() => SupportFaq.fromJson(raw);
}

class SupportTicketModel {
  const SupportTicketModel({required this.raw});
  final Map<String, dynamic> raw;
  factory SupportTicketModel.fromJson(Map<String, dynamic> json) => SupportTicketModel(raw: json);
  SupportTicket toEntity() => SupportTicket.fromJson(raw);
}

class SupportAssistantMessageModel {
  const SupportAssistantMessageModel({required this.raw});
  final Map<String, dynamic> raw;
  factory SupportAssistantMessageModel.fromJson(Map<String, dynamic> json) =>
      SupportAssistantMessageModel(raw: json);
  SupportAssistantMessage toEntity() => SupportAssistantMessage.fromJson(raw);
}
