import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/support_entity.dart';
import '../models/support_model.dart';

class SupportRemoteDataSource {
  const SupportRemoteDataSource(this._client);
  final ApiClient _client;

  static const _faqPath = '/support/faq';
  static const _ticketsPath = '/support/tickets';
  static const _assistantPath = '/support/assistant/message';

  Future<Result<List<SupportFaq>>> fetchFaq() async {
    return _client.get<List<SupportFaq>>(
      _faqPath,
      parser: (json) => (json as List<dynamic>)
          .map((e) => SupportFaqModel.fromJson(e as Map<String, dynamic>).toEntity())
          .toList(),
    );
  }

  Future<Result<List<SupportTicket>>> fetchTickets({Map<String, dynamic>? params}) async {
    return _client.get<List<SupportTicket>>(
      _ticketsPath,
      queryParameters: params,
      parser: (json) => (json as List<dynamic>)
          .map((e) => SupportTicketModel.fromJson(e as Map<String, dynamic>).toEntity())
          .toList(),
    );
  }

  Future<Result<SupportTicket>> fetchTicket(String id) async {
    return _client.get<SupportTicket>(
      '$_ticketsPath/$id',
      parser: (json) => SupportTicketModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  Future<Result<SupportTicket>> createTicket({
    required String subject,
    required String body,
    String? orderId,
    String? category,
    String? idempotencyKey,
  }) async {
    return _client.post<SupportTicket>(
      _ticketsPath,
      data: {
        'subject': subject,
        'body': body,
        if (orderId != null) 'order_id': orderId,
        if (category != null) 'category': category,
      },
      idempotencyKey: idempotencyKey,
      parser: (json) => SupportTicketModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  Future<Result<SupportAssistantMessage>> sendAssistantMessage({
    required String message,
    String? sessionId,
    String? orderId,
  }) async {
    return _client.post<SupportAssistantMessage>(
      _assistantPath,
      data: {
        'message': message,
        if (sessionId != null) 'session_id': sessionId,
        if (orderId != null) 'order_id': orderId,
      },
      parser: (json) => SupportAssistantMessageModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }
}
